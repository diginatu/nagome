package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/diginatu/nagome/nicolive"
)

const (
	eventBufferSize = 5
	accountFileName = "account.yml"
	logFileName     = "info.log"
)

var (
	printVersion  bool
	printHelp     bool
	debugToStderr bool
	standAlone    bool
	uiUseTCP      bool
	tcpPort       string
)

func init() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath",
		findUserConfigPath(), "Set <string> to save directory.")
	flag.BoolVar(&printHelp, "help", false, "Print this help.")
	flag.BoolVar(&printHelp, "h", false, "Print this help. (shorthand)")
	flag.BoolVar(&printVersion, "v", false, "Print version information.")
	flag.BoolVar(&debugToStderr, "dbgtostd", false,
		`Output debug information to stderr.
	(in default, output to the log file in the save directory)`)
	flag.BoolVar(&standAlone, "standalone", false, `Run in stand alone mode (CUI).`)
	flag.BoolVar(&uiUseTCP, "uitcp", false, `Use TCP connection for UI instead of stdin/out`)
	flag.StringVar(&tcpPort, "p", "8025", `Port to wait TCP server for UI. (see uitcp)`)
}

// RunCli processes flags and io
func RunCli() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	mkplug := flag.String("makeplug", "", "Make new plugin template with given name.")

	flag.Parse()

	pluginPath := filepath.Join(App.SavePath, "plugin")

	if printHelp {
		flag.Usage()
		os.Exit(0)
	}
	if printVersion {
		fmt.Println(App.Name, " ", App.Version)
		os.Exit(0)
	}
	if *mkplug != "" {
		p := filepath.Join(pluginPath, *mkplug)

		// check if the directory already exists
		_, err := os.Stat(p)
		if err == nil {
			log.Fatalln("Same name of plugin directory is already exists.")
		}
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}

		if err := os.MkdirAll(p, 0777); err != nil {
			log.Fatalln("could not make save directory\n", err)
		}

		pl := plugin{
			Name:    *mkplug,
			Version: "1.0",
			Depends: []string{DomainNagome},
			Method:  "tcp",
			Exec:    fmt.Sprintf("./%s {{port}} {{num}}", *mkplug),
		}
		pl.savePlugin(filepath.Join(p, "plugin.yml"))

		fmt.Printf("Create your plugin in : %s\n", p)
		os.Exit(0)
	}

	if err := os.MkdirAll(pluginPath, 0777); err != nil {
		log.Fatal("could not make save directory\n", err)
	}

	// set log
	var file *os.File
	if debugToStderr {
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

	if standAlone {
		standAloneMode()
	} else {
		clientMode()
	}

}

func standAloneMode() {
	stdinReader := bufio.NewReader(os.Stdin)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, accountFileName))

	var l nicolive.LiveWaku

	for {
		fmt.Println("input broad URL or ID (or empty to quit) :")
		brdtx, err := stdinReader.ReadString('\n')
		if err != nil || brdtx == "\n" {
			return
		}
		brdtx = brdtx[:len(brdtx)-1]

		brdRg := regexp.MustCompile("(lv|co)\\d+")
		broadMch := brdRg.FindString(brdtx)
		if broadMch == "" {
			fmt.Println("invalid text")
			continue
		}

		l = nicolive.LiveWaku{Account: &ac, BroadID: broadMch}
		break
	}

	nicoerr := l.FetchInformation()
	if nicoerr != nil {
		log.Fatalln(nicoerr)
	}

	commconn := nicolive.NewCommentConnection(&l, nil)
	commconn.Connect()

	for {
		text, err := stdinReader.ReadString('\n')
		if err != nil {
			commconn.Disconnect()
			return
		}
		text = text[:len(text)-1]

		switch text {
		case ":q":
			commconn.Disconnect()
			return
		default:
			commconn.SendComment(text, false)
		}
	}
}

func clientMode() {
	var plugs []*plugin

	// add main plugin
	var plug *plugin
	if uiUseTCP {
		plug = &plugin{
			Name:        pluginNameMain,
			Description: "main plugin (UI with TCP connection)",
			Version:     "0.0",
			Depends:     []string{DomainNagome},
		}
		plug.Init(1)
	} else {
		plug = &plugin{
			Name:        pluginNameMain,
			Description: "main plugin (UI with stdin/out connection)",
			Version:     "0.0",
			Depends:     []string{DomainNagome},
			Rw: bufio.NewReadWriter(
				bufio.NewReader(os.Stdin),
				bufio.NewWriter(os.Stdout)),
		}
		plug.Init(1)
		plug.Enable()
	}
	plugs = append(plugs, plug)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, accountFileName))

	var l nicolive.LiveWaku
	var cv = commentViewer{
		Ac:   &ac,
		Pgns: plugs,
		Evch: make(chan *Message, eventBufferSize),
		Quit: make(chan struct{}),
	}
	eventReceiver := &commentEventEmit{cv: &cv}
	cv.Cmm = nicolive.NewCommentConnection(&l, eventReceiver)

	cv.Run()
}
