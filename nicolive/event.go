package nicolive

import "fmt"

// EventClassNum is an Enum to represent Event.Class
type EventClassNum int

// EventClass
const (
	EventClassCommentConnection EventClassNum = iota
	EventClassPlayerStatus
	EventClassHeartBeat
)

// EventTypeNum is an Enum to represent Event.Type
type EventTypeNum int

// EventType
const (
	EventTypeErr EventTypeNum = iota
	EventTypeGot
	EventTypeWakuEnd
	EventTypeSend
	EventTypeOpen
	EventTypeClose
)

// Event is an event
type Event struct {
	Class   EventClassNum
	Type    EventTypeNum
	ID      string
	Content interface{}
}

// MakeEventID makes a string to represent event ID
func MakeEventID(mail, broadID string) string {
	return fmt.Sprintf("%s:%s", mail, broadID)
}

// IsEqID returns whether same ID or not
func (e *Event) IsEqID(mail, broadID string) bool {
	return e.ID == MakeEventID(mail, broadID)
}

func (e *Event) String() string {
	var cls string
	switch e.Class {
	case EventClassCommentConnection:
		cls = "CommentConnection"
	case EventClassPlayerStatus:
		cls = "PlayerStatus"
	case EventClassHeartBeat:
		cls = "HeartBeat"
	}
	var tys string
	switch e.Type {
	case EventTypeErr:
		tys = "error"
	case EventTypeGot:
		tys = "got"
	case EventTypeWakuEnd:
		tys = "waku ended"
	case EventTypeSend:
		tys = "send"
	case EventTypeOpen:
		tys = "open"
	case EventTypeClose:
		tys = "close"
	}
	return fmt.Sprintf("%s.%s %s", cls, tys, e.Content)
}

// EventReceiver receive events and proceed
type EventReceiver interface {
	Proceed(*Event)
}

type defaultEventReceiver struct{}

func (der *defaultEventReceiver) Proceed(ev *Event) {
	Logger.Println(ev)
}
