package nicolive

import "fmt"

// EventTypeNum is an Enum to represent Event.Type
type EventTypeNum int

// EventType
const (
	EventTypeErr EventTypeNum = iota
	EventTypeGot
	EventTypeSend
	EventTypeOpen
	EventTypeClose
	EventTypeWakuEnd
	EventTypeHeartBeatGot
	EventTypeAntennaOpen
	EventTypeAntennaClose
	EventTypeAntennaGot
	EventTypeAntennaErr
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
		tys = "Err"
	case EventTypeGot:
		tys = "Got"
	case EventTypeWakuEnd:
		tys = "WakuEnd"
	case EventTypeSend:
		tys = "Send"
	case EventTypeOpen:
		tys = "Open"
	case EventTypeClose:
		tys = "Close"
	case EventTypeHeartBeatGot:
		tys = "HeatBeatGot"
	case EventTypeAntennaOpen:
		tys = "AntennaOpen"
	case EventTypeAntennaClose:
		tys = "AntennaClose"
	case EventTypeAntennaGot:
		tys = "AntennaGot"
	case EventTypeAntennaErr:
		tys = "AntennaErr"
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
