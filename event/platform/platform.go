package platform

import (
	"github.com/micro/go-os/event"
)

func NewEvent(opts ...event.Option) event.Event {
	return event.NewEvent(opts...)
}
