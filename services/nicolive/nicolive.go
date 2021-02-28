// Package nicolive provides access way to NicoNama API,
// corresponding structures and other features.
package nicolive

import (
	"fmt"
	"runtime"

	"gopkg.in/xmlpath.v2"
)

var (
	xmlPathStatus    = xmlpath.MustCompile("//@status")
	xmlPathTime      = xmlpath.MustCompile("//@time")
	xmlPathErrorCode = xmlpath.MustCompile("//error/code")
	xmlPathErrorDesc = xmlpath.MustCompile("//error/description")
)

func caller(sk int) string {
	_, file, line, ok := runtime.Caller(sk)
	short := file
	if !ok {
		short = "???"
		line = 0
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", short, line)
}
