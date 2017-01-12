package main

import (
	"fmt"
	"os"
)

// Application holds app settings and valuables
type Application struct {
	SavePath      string
	SettingsSlots SettingsSlots
}

// Application global information
var (
	AppName = "Nagome"
	Version string
	App     Application
)

func main() {
	if Version == "" {
		fmt.Fprintln(os.Stderr, "Version value was not set at build time.")
		Version = "Unknown"
	}
	RunCli()
	return
}
