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
	eventBufferSize = 100
	accountFileName = "account.yml"
	logFileName     = "info.log"
	pluginDirName   = "plugin"
)

// RunCli processes flags and io
func RunCli() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath",
		findUserConfigPath(), "Set <string> to save directory.")
	tcpPort := flag.String("p", "8025", `Port to wait TCP server for UI. (see uitcp)`)
	debugToStderr := flag.Bool("dbgtostd", false,
		`Output debug information to stderr.
	(in default, output to the log file in the save directory)`)
	standAlone := flag.Bool("standalone", false, `Run in stand alone mode (CUI).`)
	uiUseTCP := flag.Bool("uitcp", false, `Use TCP connection for UI instead of stdin/out`)
	printHelp := flag.Bool("help", false, "Print this help.")
	printHelp = flag.Bool("h", false, "Print this help. (shorthand)")
	printVersion := flag.Bool("v", false, "Print version information.")

	log.SetFlags(log.Lshortfile | log.Ltime)
	mkplug := flag.String("makeplug", "", "Make new plugin template with given name.")

	flag.Parse()

	pluginPath := filepath.Join(App.SavePath, pluginDirName)

	if *printHelp {
		flag.Usage()
		os.Exit(0)
	}
	if *printVersion {
		fmt.Println(App.Name, " ", App.Version)
		os.Exit(0)
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
		pl.savePlugin(filepath.Join(p, "plugin.yml"))

		fmt.Printf("Create your plugin in : %s\n", p)
		os.Exit(0)
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

	if *standAlone {
		standAloneMode()
	} else {
		clientMode(*uiUseTCP, tcpPort)
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
		log.Fatalln(nicoerr.Error())
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

func clientMode(uiUseTCP bool, tcpPort *string) {
	var cv = CommentViewer{
		Ac:      new(nicolive.Account),
		TCPPort: *tcpPort,
		Evch:    make(chan *Message, eventBufferSize),
		Quit:    make(chan struct{}),
	}
	cv.Cmm = nicolive.NewCommentConnection(new(nicolive.LiveWaku), &cv)

	// load account data
	cv.Ac.Load(filepath.Join(App.SavePath, accountFileName))

	// add main plugin
	var plug *plugin
	if uiUseTCP {
		plug = &plugin{
			Name:        pluginNameMain,
			Description: "main plugin (UI with TCP connection)",
			Version:     "0.0",
			Depends:     []string{DomainNagome, DomainComment, DomainUI},
		}
		plug.Init(1)
	} else {
		plug = &plugin{
			Name:        pluginNameMain,
			Description: "main plugin (UI with stdin/out connection)",
			Version:     "0.0",
			Depends:     []string{DomainNagome, DomainComment, DomainUI},
			Rw:          bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)),
		}
		plug.Init(1)
		plug.Start(&cv)
	}
	cv.Pgns = append(cv.Pgns, plug)

	cv.Run()
}
