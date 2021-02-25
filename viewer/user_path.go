// +build !windows

package viewer

import (
	"os"
	"path/filepath"
	"strings"
)

func findUserConfigPath(appname string) string {
	xdgConf := os.Getenv("XDG_CONFIG_HOME")
	if xdgConf == "" {
		home := os.Getenv("HOME")
		xdgConf = filepath.Join(home, ".config")
	}

	dir := filepath.Join(xdgConf, strings.ToLower(appname))

	return dir
}
