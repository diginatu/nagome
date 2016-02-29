package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	// Logger is logger in this app
	Logger        *log.Logger
	printVersion  bool
	printHelp     bool
	debugToStderr bool
)

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
}

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

	return
}
