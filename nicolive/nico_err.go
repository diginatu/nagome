package nicolive

import (
	"fmt"
)

// ErrNum is an Enum to represent error number to identify.
type ErrNum int

// Enum ErrNum
const (
	ErrOther ErrNum = iota
	ErrSendComment
	ErrConnection
	ErrNicoLiveOther
	ErrNotLogin
	ErrClosed
	ErrIncorrectAccount
)

// Error is error interface in nicolive
type Error interface {
	No() ErrNum          // for identifying by application  (ie. select messages to the user)
	Description() string // messages for debug information
	Where() string       // where the error occurred
	Error() string
}

// ErrStruct is an error struct in nicolive
type ErrStruct struct {
	no    ErrNum
	desc  string
	where string
}

func (n ErrStruct) Error() string {
	return fmt.Sprintf("[%s] %s", n.where, n.desc)
}

// No returns errorNum
func (n ErrStruct) No() ErrNum {
	return n.no
}

// Description returns description
func (n ErrStruct) Description() string {
	return n.desc
}

// Where returns where
func (n ErrStruct) Where() string {
	return n.where
}

// MakeError returns Error that format as the given info
// and the code position.
func MakeError(errNum ErrNum, description string) Error {
	where := caller(2)
	return &ErrStruct{errNum, description, where}
}

// ErrFromStdErr returns Error converted from standard error.
func ErrFromStdErr(e error) Error {
	where := caller(2)
	return &ErrStruct{ErrOther, e.Error(), where}
}
