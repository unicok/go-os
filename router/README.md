# Router interface

The router is a client interface to a global or federated regional router. 
A go-micro client uses the registry and then on an individual basis calculates 
how to route. The router allows decisions to be made across the platform.

## Interface

```go
// The router is the client interface to a
// global service loadbalancer (GSLB).
// Metrics are batched and published to
// a router which has a view of the whole
// system.
type Router interface {
	// Provides the selector interface
	selector.Selector
	// Return stats maintained here
	Stats() ([]*Stats, error)
	// Record stats for a request - too many args ugh
	Record(r Request, node *registry.Node, d time.Duration, err error)
	// Start/Stop internal publishing, caching, etc
	Start() error
	Stop() error
}

type Stats struct {
	Service   *registry.Service
	Client    *registry.Service
	Timestamp int64
	Duration  int64
	// TODO:
	// Selected
	// Endpoints
}

func NewRouter(opts ...selector.Option) Router {
	return newPlatform(opts...)
}
```

## Supported Backends

- Router service
