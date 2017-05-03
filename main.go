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

	cli := &CLI{
		InStream:  os.Stdin,
		OutStream: os.Stdout,
		ErrStream: os.Stderr,
	}

	os.Exit(cli.RunCli(os.Args))
	return
}
