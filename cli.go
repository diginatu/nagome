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
}

func runCui() {
	stdinReader := bufio.NewReader(os.Stdin)

	var ac nicolive.Account
	ac.Load(filepath.Join(App.SavePath, "userData.yml"))

	//err := ac.Login()
	//if err != nil {
	//Logger.Fatalln(err)
	//}
	//ac.Save(filepath.Join(App.SavePath, "userData.yml"))

	var l nicolive.LiveWaku

	for {
		fmt.Println("input broad URL or ID :")
		brdtx, err := stdinReader.ReadString('\n')
		brdtx = brdtx[:len(brdtx)-1]
		if err != nil || brdtx == "" {
			return
		}

		brdRg := regexp.MustCompile("lv(\\d+)")
		broadMch := brdRg.FindString(brdtx)
		if err != nil {
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

	commconn := nicolive.NewCommentConnection(&l)
	commconn.Connect()

	for {
		text, err := stdinReader.ReadString('\n')
		text = text[:len(text)-1]
		if err != nil || text == ":q" {
			commconn.Disconnect()
			return
		}

		if text != "" {
			commconn.SendComment(text, false)
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

	//mainProcess()
	runCui()
}
