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
	ErrOpen
	ErrConnection
	ErrNicoLiveOther
	ErrNotLogin
	ErrClosed
	ErrRequireCommunityMember
	ErrIncorrectAccount
	ErrNetwork
	ErrDBUserNotFound
)

// Error is an error struct in nicolive
type Error struct {
	etype ErrNum
	desc  string
	where string
}

// TypeString returns name of the error type.
func (e Error) TypeString() string {
	var s string
	switch e.etype {
	case ErrOther:
		s = "other"
	case ErrSendComment:
		s = "sending comment"
	case ErrConnection:
		s = "connection"
	case ErrNicoLiveOther:
		s = "nico live other"
	case ErrNotLogin:
		s = "not login"
	case ErrClosed:
		s = "closed live"
	case ErrRequireCommunityMember:
		s = "require_community_member"
	case ErrIncorrectAccount:
		s = "incorrect account"
	}
	return s
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] <%s> %s", e.where, e.TypeString(), e.desc)
}

// Type returns errorNum for identifying by application.
// (ie. select messages to the user)
func (e Error) Type() ErrNum {
	return e.etype
}

// Description returns description
func (e Error) Description() string {
	return e.desc
}

// Where returns where
func (e Error) Where() string {
	return e.where
}

// MakeError returns Error that is formated as the given info and the code position.
func MakeError(errNum ErrNum, description string) Error {
	where := caller(2)
	return Error{errNum, description, where}
}

// ErrFromStdErr returns Error converted from standard error.
func ErrFromStdErr(e error) Error {
	where := caller(2)
	return Error{ErrOther, e.Error(), where}
}
