package viewer

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/diginatu/nagome/nicolive"
)

const (
	eventBufferSize  = 50
	accountFileName  = "account.yml"
	logFileName      = "info.log"
	logFlags         = log.Lshortfile | log.Ltime
	pluginDirName    = "plugin"
	pluginConfigName = "plugin.yml"
	userDBDirName    = "userdb"
	settingsFileName = "setting.yml"
)

// CLI has valuables and settings for a CLI environment.
type CLI struct {
	InStream             io.ReadCloser
	OutStream, ErrStream io.WriteCloser
	SavePath             string
	SettingsSlots        SettingsSlots
	log                  *log.Logger
	AppName, Version     string
}

// NewCLI creates new default values CLI struct.
func NewCLI(name, appname string) *CLI {
	return &CLI{
		InStream:  os.Stdin,
		OutStream: os.Stdout,
		ErrStream: os.Stderr,
		AppName:   appname,
		SavePath:  findUserConfigPath(appname),
		log:       log.New(os.Stderr, name, logFlags),
	}
}

// RunCli runs CLI functions as one command line program.
// This returns the CLI return value.
func (c *CLI) RunCli(args []string) int {
	// set command line options
	var (
		printHelp bool
		mainyml   string
	)

	flagst := flag.NewFlagSet("nagome-cli", flag.ContinueOnError)
	flagst.SetOutput(c.ErrStream)

	flagst.StringVar(&c.SavePath, "savepath", findUserConfigPath(c.AppName), "Set <string> to save directory.")
	tcpPort := flagst.String("p", "8025", `Port to wait TCP server for plugins.  Set 0 to try to find free port.  (see docs/plugin.md)`)
	debugToStderr := flagst.Bool("dbgtostd", false, `Output debug information to stderr.
	(in default, output to the log file in the save directory)`)
	flagst.BoolVar(&printHelp, "help", false, "Print this help.")
	flagst.BoolVar(&printHelp, "h", false, "Print this help. (shorthand)")
	printVersion := flagst.Bool("v", false, "Print version information.")
	mkplug := flagst.String("makeplug", "", "Make new plugin template with given name.")
	flagst.StringVar(&mainyml, "ymlmain", "", `specfy the config file of main plugin.
	Its format is same as yml file of normal plugins.`)
	flagst.StringVar(&mainyml, "y", "", `specfy the config file of main plugin. (shorthand)`)

	err := flagst.Parse(args[1:])
	if err != nil {
		return 1
	}

	// set log
	var logw io.Writer
	if *debugToStderr {
		logw = c.ErrStream
	} else {
		file, err := os.Create(filepath.Join(c.SavePath, logFileName))
		if err != nil {
			c.log.Println("could not open log file\n", err)
			return 1
		}
		defer func() {
			err := file.Close()
			if err != nil {
				c.log.Println(err)
			}
		}()
		logw = file
	}
	c.log.SetOutput(logw)

	err = c.SettingsSlots.Load(filepath.Join(c.SavePath, settingsFileName))
	if err != nil {
		c.log.Println(err)
	}

	pluginPath := filepath.Join(c.SavePath, pluginDirName)

	if printHelp {
		flagst.Usage()
		return 0
	}
	if *printVersion {
		fmt.Fprintln(c.OutStream, c.AppName, " ", c.Version)
		return 0
	}
	if *mkplug != "" {
		err = c.generatePluginTemplate(*mkplug, pluginPath)
		if err != nil {
			c.log.Println(err)
			return 1
		}
		return 0
	}

	if err := os.MkdirAll(pluginPath, 0777); err != nil {
		c.log.Println("could not make save directory\n", err)
		return 1
	}

	cv := NewCommentViewer(*tcpPort, c)

	ac, err := nicolive.AccountLoad(filepath.Join(c.SavePath, accountFileName))
	if err != nil {
		c.log.Println(err)
		cv.Ac = new(nicolive.Account)
		cv.Evch <- NewMessageMust(DomainUI, CommUIConfigAccount, nil)
	} else {
		cv.Ac = ac
	}

	// add main plugin
	plug := newPlugin(cv)
	plug.Name = "main"
	plug.Description = "main plugin"
	plug.Version = "0.0"
	plug.Method = pluginMethodStd
	plug.Subscribe = []string{DomainNagome, DomainComment, DomainUI}
	if mainyml != "" {
		err = plug.Load(mainyml)
		if err != nil {
			c.log.Println(err)
			return 1
		}
	}
	cv.AddPlugin(plug)
	if plug.Method == pluginMethodStd {
		err := plug.Open(&stdReadWriteCloser{c.InStream, c.OutStream}, true)
		if err != nil {
			c.log.Println(err)
			return 1
		}
	}

	cv.Start()
	cv.Wait()

	if cv.Settings.AutoSaveTo0Slot {
		c.SettingsSlots.Config[0] = &cv.Settings
	}
	err = c.SettingsSlots.Save(filepath.Join(cv.cli.SavePath, settingsFileName))
	if err != nil {
		c.log.Println(err)
		return 1
	}

	return 0
}

func (c *CLI) generatePluginTemplate(name, pluginPath string) error {
	p := filepath.Join(pluginPath, name)

	// check if the directory already exists
	_, err := os.Stat(p)
	if err == nil {
		return fmt.Errorf("same name directory is already exists")
	}

	if err := os.MkdirAll(p, 0777); err != nil {
		return fmt.Errorf("Could not make save directory : %s", err)
	}

	pl := Plugin{
		Name:      name,
		Version:   "1.0",
		Subscribe: []string{DomainNagome},
		Method:    "tcp",
		Exec:      []string{"{{path}}/" + name, "{{port}}", "{{no}}"},
	}
	err = pl.Save(filepath.Join(p, pluginConfigName))
	if err != nil {
		return fmt.Errorf("failed to save file : %s", err)
	}

	fmt.Fprintf(c.OutStream, "Create your plugin in : %s\n", p)
	return nil
}
