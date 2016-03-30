package platform

import (
	"github.com/micro/go-platform/event"
)

func NewEvent(opts ...event.Option) event.Event {
	return event.NewEvent(opts...)
}
