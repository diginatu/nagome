package main

import (
	"fmt"
	"os"
)

// Application global information
var (
	AppName = "Nagome"
	Version string
)

func main() {
	if Version == "" {
		fmt.Fprintln(os.Stderr, "Version value was not set at build time.")
		Version = "Unknown"
	}

	cli := NewCLI("")
	os.Exit(cli.RunCli(os.Args))
	return
}
