package nicolive

import (
	"fmt"
	"runtime"
)

// NicoErrNum is an Enum to represent Nico.Errnum
type NicoErrNum int

// Enum NicoErrNum
const (
	NicoErrOther NicoErrNum = iota - 1
	NicoErrNicoLiveOther
	NicoErrNotLogin
)

// NicoErr is an error struct in nicolive
type NicoErr struct {
	Errnum      NicoErrNum
	Code        string
	Description string
	Where       string
}

func (n NicoErr) Error() string {
	if n.Description == "" {
		return fmt.Sprintf("[%s] %s", n.Where, n.Code)
	}
	return fmt.Sprintf("[%s] %s : %s", n.Where, n.Code, n.Description)
}

// NewNicoErr returns NicoErr that format as the given info
// and the code position.
func NewNicoErr(errNum NicoErrNum, code string, description string) *NicoErr {
	_, file, line, ok := runtime.Caller(1)
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
	where := fmt.Sprintf("%s:%d", short, line)

	return &NicoErr{errNum, code, description, where}
}

// NewNicoErrFromStdErr returns NicoErr converted from standard error.
func NewNicoErrFromStdErr(e error) *NicoErr {
	return NewNicoErr(NicoErrOther, "standard error", e.Error())
}
