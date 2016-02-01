package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/diginatu/Nagome/nicolive"
)

// Application holds app settings and valuables
type Application struct {
	// Name is name of this app
	Name string
	// Version is version info
	Version string
	// SavePath is directory to hold save files
	SavePath string
}

var (
	// App is global Application settings and valuables for this app
	App = Application{Name: "Nagome", Version: "0.0"}
	// Logger is logger in this app
	Logger        *log.Logger
	printVersion  bool
	printHelp     bool
	debugToStderr bool
)

func main() {
	flag.Parse()

	if printHelp {
		flag.Usage()
		return
	}
	if printVersion {
		fmt.Println(App.Name, " ", App.Version)
		return
	}

	err := os.MkdirAll(App.SavePath, 0777)
	if err != nil {
		log.Fatal("could not make save directory\n" + err.Error())
	}

	var file *os.File
	if debugToStderr {
		file = os.Stderr
	} else {
		file, err = os.Create(filepath.Join(App.SavePath, "info.log"))
		if err != nil {
			log.Fatal("could not open log file\n" + err.Error())
		}
	}
	defer file.Close()
	Logger = log.New(file, "", log.Lshortfile|log.Ltime)

	// below is test code
	fmt.Println("Hello ", App.Name)
	var ac nicolive.Account
	ac.LoadAccount(filepath.Join(App.SavePath, "userData.yml"))

	l := nicolive.LiveWaku{Account: &ac, BroadID: "lv1234567"}
	err = l.FetchInformation()
	if err != nil {
		Logger.Fatalln(err)
	}

	return
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

func findUserConfigPath() string {
	var home, dir string

	switch runtime.GOOS {
	case "windows":
		home = os.Getenv("USERPROFILE")
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(home, "Application Data")
		}
	case "plan9":
		home = os.Getenv("home")
		dir = filepath.Join(home, ".config")
	default:
		home = os.Getenv("HOME")
		dir = filepath.Join(home, ".config")
	}

	return filepath.Join(dir, App.Name)
}
