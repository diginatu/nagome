package nicolive

import (
	"fmt"
	"runtime"
)

// NicoErrNum is Enum to represent Errnum
type NicoErrNum int

// Enum NicoErrNum
const (
	Other    NicoErrNum = iota // Other is other NicoErrNum
	NotLogin                   // NotLogin is that user is not login
)

// NicoErr is an error struct in nicolive
type NicoErr struct {
	Errnum      NicoErrNum
	Code        string
	Description string
	Where       string
}

func (n NicoErr) Error() string {
	return fmt.Sprintf("[%s] %s : %s", n.Where, n.Code, n.Description)
}

// NewNicoErr returns NicoErr that format as the given info
func NewNicoErr(errNum NicoErrNum, code string, description string) NicoErr {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short

	return NicoErr{errNum, code, description, fmt.Sprintf("%s:%d", file, line)}
}
