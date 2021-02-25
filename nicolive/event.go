package nicolive

import "fmt"

// EventTypeNum is an Enum to represent Event.Type
type EventTypeNum int

// EventType
const (
	EventTypeCommentErr EventTypeNum = iota
	EventTypeCommentGot
	EventTypeCommentSend
	EventTypeCommentOpen
	EventTypeCommentClose
	EventTypeWakuEnd
	EventTypeHeartBeatGot
)

// Event is an event
type Event struct {
	Type    EventTypeNum
	Content interface{}
}

func (e *Event) String() string {
	var tys string
	switch e.Type {
	case EventTypeCommentErr:
		tys = "CommentErr"
	case EventTypeCommentGot:
		tys = "CommentGot"
	case EventTypeWakuEnd:
		tys = "WakuEnd"
	case EventTypeCommentSend:
		tys = "CommentSend"
	case EventTypeCommentOpen:
		tys = "CommentOpen"
	case EventTypeCommentClose:
		tys = "CommentClose"
	case EventTypeHeartBeatGot:
		tys = "HeatBeatGot"
	}
	return fmt.Sprintf("Event {Type:%s %s}", tys, e.Content)
}

// EventReceiver receive events and proceed
type EventReceiver interface {
	ProceedNicoEvent(*Event)
}

type defaultEventReceiver struct{}

func (der defaultEventReceiver) ProceedNicoEvent(ev *Event) {
	fmt.Println(caller(3), ":", ev)
}
