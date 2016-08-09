package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/diginatu/nagome/nicolive"
)

const (
	eventBufferSize = 5
)

var (
	// Logger is logger in this app
	Logger        *log.Logger
	printVersion  bool
	printHelp     bool
	debugToStderr bool
	standAlone    bool
	runBgproc     bool
)

func init() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath",
		findUserConfigPath(), "Set <directory> to save directory.")
	flag.BoolVar(&printHelp, "h", false, "Print this help.")
	flag.BoolVar(&printHelp, "help", false, "Print this help.")
	flag.BoolVar(&printVersion, "version", false, "Print version information.")
	flag.BoolVar(&debugToStderr, "dbgtostd", false,
		`Output debug information to stderr.
	Without this option, output to the log file in the save directory.`)
	flag.BoolVar(&standAlone, "standalone", false, `Run in stand alone mode (CUI).`)
	flag.BoolVar(&runBgproc, "bg", false, `Run as daemon. (use stdin/out as one connection to a plugin)`)
}

// RunCli processes flags and io
func RunCli() {
	flag.Parse()

	if printHelp {
		flag.Usage()
		os.Exit(0)
	}
	if printVersion {
		fmt.Println(App.Name, " ", App.Version)
		os.Exit(0)
	}

	if err := os.MkdirAll(App.SavePath, 0777); err != nil {
		log.Fatal("could not make save directory\n" + err.Error())
	}

	var file *os.File
	if debugToStderr {
		file = os.Stderr
	} else {
		var err error
		file, err = os.Create(filepath.Join(App.SavePath, "info.log"))
		if err != nil {
			log.Fatal("could not open log file\n" + err.Error())
		}
	}
	defer file.Close()
	Logger = log.New(file, "", log.Lshortfile|log.Ltime)

	if standAlone {
		standAloneMode()
		return
	}

	clientMode()
}

func standAloneMode() {
	stdinReader := bufio.NewReader(os.Stdin)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))

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
		Logger.Fatalln(nicoerr)
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

	if runBgproc {
		plug := &plugin{
			Name:        pluginNameMain,
			Description: "main plugin(UI)",
			Version:     "0.0",
			Depends:     []string{DomainNagome},
			Rw: bufio.NewReadWriter(
				bufio.NewReader(os.Stdin),
				bufio.NewWriter(os.Stdout)),
			FlushTm: time.NewTimer(time.Hour),
		}
		plugs = append(plugs, plug)
	}

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))

	var l nicolive.LiveWaku
	var cmvw = commentViewer{
		Ac:   &ac,
		Cmm:  nicolive.NewCommentConnection(&l, nil),
		Pgns: plugs,
		Evch: make(chan *Message, eventBufferSize),
		Quit: make(chan struct{}),
	}

	cmvw.runCommentViewer()
}
