// +build !windows

package main

import (
	"os"
	"path/filepath"
)

func findUserConfigPath() string {
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".config")

	return filepath.Join(dir, App.Name)
}
