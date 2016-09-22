package main

import (
	"os"
	"path/filepath"
)

func findUserConfigPath() string {
	home := os.Getenv("USERPROFILE")
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir = filepath.Join(home, "Application Data")
	}

	return filepath.Join(dir, App.Name)
}
