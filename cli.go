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

var (
	// Logger is logger in this app
	Logger        *log.Logger
	printVersion  bool
	printHelp     bool
	debugToStderr bool
)

func init() {
	// set command line options
	flag.StringVar(&App.SavePath, "savepath",
		findUserConfigPath(), "Set <directory> to save directory.")
	flag.BoolVar(&printHelp, "h", false, "Print this help.")
	flag.BoolVar(&printHelp, "help", false, "Print this help.")
	flag.BoolVar(&printVersion, "version", false, "Print version information.")
	flag.BoolVar(&debugToStderr, "dbgtostd", false,
		"Output debug info into stderr\n"+
			"otherwise save to the log file in the save directory.")
}

func mainProcess() {
	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))
	//ac.Save(filepath.Join(App.SavePath, "userData.yml"))

	//err = ac.Login()
	//if err != nil {
	//Logger.Fatalln(err)
	//}

	l := nicolive.LiveWaku{Account: &ac, BroadID: "lv256606827"}
	nicoerr := l.FetchInformation()
	if nicoerr != nil {
		Logger.Fatalln(nicoerr)
	}

	commentconn := nicolive.NewCommentConnection(&l)
	commentconn.Connect()

	stdinReader := bufio.NewReader(os.Stdin)

	for {
		text, err := stdinReader.ReadString('\n')
		if err != nil || text == "close\n" {
			commentconn.Disconnect()
			return
		}
	}
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

	mainProcess()
}
