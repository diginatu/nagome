package main

import (
	"fmt"
	"os"

	"github.com/diginatu/nagome/viewer"
)

const (
	// AppName is the application name
	AppName = "Nagome"
)

// Application global information
var (
	Version string
)

func main() {
	if Version == "" {
		fmt.Fprintln(os.Stderr, "Version value was not set at build time.")
		Version = "Unknown"
	}

	cli := viewer.NewCLI("", AppName)
	cli.Version = Version
	os.Exit(cli.RunCli(os.Args))
	return
}
