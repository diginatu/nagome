package nicolive

import "fmt"

// EventClassNum is an Enum to represent Event.Class
type EventClassNum int

// EventClass
const (
	EventClassCommentConnection EventClassNum = iota
)

// EventTypeNum is an Enum to represent Event.Type
type EventTypeNum int

// EventType
const (
	EventTypeErr EventTypeNum = iota
)

// Event is an event
type Event struct {
	Class   EventClassNum
	Type    EventTypeNum
	ID      int
	Content interface{}
}

func (e *Event) String() string {
	var cls string
	switch e.Class {
	case EventClassCommentConnection:
		cls = "CommentConnection"
	}
	var tys string
	switch e.Type {
	case EventTypeErr:
		tys = "error"
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
