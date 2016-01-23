package event

type Event interface {
	// publish an event record
	Publish(*Record) error
	// subscribe to an event types
	Subscribe(Handler, ...string) error
	// used for internal purposes
	Start() error
	Stop() error
}

type Record struct {
	Id        string
	Type      string
	Origin    string
	Timestamp int64
	RootId    string
	Metadata  map[string]string
	Data      string
}

type Handler func(*Record)

var (
	RecordTopic      = "platform.event.record"
	DefaultEventType = "event"
)

type Option func(o *Options)

func NewEvent(opts ...Option) Event {
	return newPlatform(opts...)
}
