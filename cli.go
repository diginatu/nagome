package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diginatu/nagome/nicolive"
)

const (
	eventBufferSize  = 50
	accountFileName  = "account.yml"
	logFileName      = "info.log"
	pluginDirName    = "plugin"
	pluginConfigName = "plugin.yml"
)

// RunCli processes flags and io
func RunCli() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath", findUserConfigPath(), "Set <string> to save directory.")
	tcpPort := flag.String("p", "8025", `Port to wait TCP server for UI. (see uitcp)`)
	debugToStderr := flag.Bool("dbgtostd", false, `Output debug information to stderr.
	(in default, output to the log file in the save directory)`)
	printHelp := flag.Bool("help", false, "Print this help.")
	printHelp = flag.Bool("h", false, "Print this help. (shorthand)")
	printVersion := flag.Bool("v", false, "Print version information.")
	mkplug := flag.String("makeplug", "", "Make new plugin template with given name.")
	mainyml := flag.String("ymlmain", "", `specfy the config file of main plugin.
	Its format is same as yml file of normal plugins.`)
	mainyml = flag.String("y", "", `specfy the config file of main plugin. (shorthand)`)

	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ltime)

	pluginPath := filepath.Join(App.SavePath, pluginDirName)

	if *printHelp {
		flag.Usage()
		return
	}
	if *printVersion {
		fmt.Println(App.Name, " ", App.Version)
		return
	}
	if *mkplug != "" {
		p := filepath.Join(pluginPath, *mkplug)

		// check if the directory already exists
		_, err := os.Stat(p)
		if err == nil {
			log.Fatalln("Same name directory is already exists.")
		}

		if err := os.MkdirAll(p, 0777); err != nil {
			log.Fatalln("could not make save directory\n", err)
		}

		pl := plugin{
			Name:    *mkplug,
			Version: "1.0",
			Depends: []string{DomainNagome},
			Method:  "tcp",
			Exec:    []string{"{{path}}/" + *mkplug, "{{port}}", "{{no}}"},
		}
		pl.savePlugin(filepath.Join(p, pluginConfigName))

		fmt.Printf("Create your plugin in : %s\n", p)
		return
	}

	if err := os.MkdirAll(pluginPath, 0777); err != nil {
		log.Fatal("could not make save directory\n", err)
	}

	// set log
	var file *os.File
	if *debugToStderr {
		file = os.Stderr
	} else {
		var err error
		file, err = os.Create(filepath.Join(App.SavePath, logFileName))
		if err != nil {
			log.Fatal("could not open log file\n", err)
		}
	}
	defer file.Close()
	log.SetOutput(file)

	cv := NewCommentViewer(new(nicolive.Account), *tcpPort)
	cv.Cmm = nicolive.NewCommentConnection(cv)

	// load account data
	err := cv.Ac.Load(filepath.Join(App.SavePath, accountFileName))
	if err != nil {
		log.Println(err)
	}

	// add main plugin
	plug := &plugin{
		Name:        "main",
		Description: "main plugin",
		Version:     "0.0",
		Method:      "std",
		Depends:     []string{DomainNagome, DomainComment, DomainUI},
	}
	if *mainyml != "" {
		err = plug.loadPlugin(*mainyml)
		if err != nil {
			log.Println(err)
		}
	}
	if plug.Method == "tcp" {
		plug.Init(1)
	} else {
		plug.Rw = bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
		plug.Init(1)
		plug.Start(cv)
	}
	cv.Pgns = append(cv.Pgns, plug)

	cv.Start()
	cv.Wait()
}
