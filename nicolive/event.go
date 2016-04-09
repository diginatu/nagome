package nicolive

import "fmt"

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
	Type    EventTypeNum
	Content interface{}
}

func (e *Event) String() string {
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
	return fmt.Sprintf("%s %s", tys, e.Content)
}

// EventReceiver receive events and proceed
type EventReceiver interface {
	Proceed(*Event)
}

type defaultEventReceiver struct{}

func (der defaultEventReceiver) Proceed(ev *Event) {
	Logger.Println(ev)
}
