package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Application holds app settings and valuables
type Application struct {
	// SavePath is directory to hold save files
	SavePath string
	// Version is version info
	Version string
}

var (
	// App is global Application for this app
	App          Application
	printVersion bool
	printHelp    bool
)

func main() {
	flag.Parse()

	if printHelp {
		flag.Usage()
		return
	}

	if printVersion {
		fmt.Println("Nagome ", App.Version)
		return
	}

	os.MkdirAll(App.SavePath, 0777)

	fmt.Println("Hello Nagome")

	return
}

func init() {
	App.Version = "0.0"

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
	defaultSavePath := filepath.Join(dir, "Nagome")

	// set command line options
	flag.StringVar(&App.SavePath, "savepath", defaultSavePath, "Set <directory> to save directory.")
	flag.BoolVar(&printHelp, "h", false, "Print this help.")
	flag.BoolVar(&printHelp, "help", false, "Print this help.")
	flag.BoolVar(&printVersion, "version", false, "Print version information.")

	return
}
