# Event interface

It's important to be able to track events at the platform level to know what's changing where. With 
hundreds of services it can be difficult to track or find a way to enforce this across services. 
By tracking platform events we can essentially build a platform event correlation system.

## Interface

```go
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
```
