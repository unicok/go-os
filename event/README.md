# Event interface

It's important to be able to track events at the platform level to know what's changing where. With 
hundreds of services it can be difficult to track or find a way to enforce this across services. 
By tracking platform events we can essentially build a platform event correlation system.

## Interface

```go

type Event interface {

	// Certain events may be registered or tracked here
	Start() error
	Stop() error
}
```
