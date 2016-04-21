# Discovery - Discovery interface

Provides a high level pluggable abstraction for discovery.

Building on ideas from [Eureka 2.0](https://github.com/Netflix/eureka/wiki/Eureka-2.0-Architecture-Overview)

## Interface

The go-micro registry provides a simplistic manner of heartbeating through re-registration, 
we want to provide heartbeating as an actual type of interaction at the discovery level. 
Discovery will provide heartbeating as a published events so anything can subscribe to 
the heartbeats. It also includes an in-memory client side cache of the registry using 
the registry Watcher. The default registry does not cache and so is limited by the 
scalabilty of the service discovery mechanism chosen. 

In the future it will also understand massive failure based on network events and stop 
from deleting the registry cache.

```go
// Discovery builds on the registry as a mechanism
// for finding services. It includes heartbeating
// to notify of liveness and caching of the registry.
type Discovery interface {
	// implements the registry interface
	registry.Registry
	// starts the watcher, caching and heartbeating
	Start() error
	// stops the watcher, caching and hearbeating
	Stop() error
}

func NewDiscovery(opts ...Option) Discovery {
	return newPlatform(opts...)
}
```

## Supported Backends

- Micro registry (any plugins; consul, etcd, memory)
- Discovery service
