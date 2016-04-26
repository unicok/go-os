# Discovery - Discovery interface

Provides a high level pluggable abstraction for discovery.

Building on ideas from [Eureka 2.0](https://github.com/Netflix/eureka/wiki/Eureka-2.0-Architecture-Overview)

## Interface

The go-micro registry provides a simple Registry abstraction for various service discovery systems. 
"Heartbeating" is also done through a simple form of re-registration. Because of this we end up 
with a system that has limited scaling potential and does not provide much information about 
service health.

Discovery provides heartbeating as events published via the Broker. This means anyone can subscribe 
to the heartbeats to determine "liveness" rather than querying the Registry. Discovery also 
includes an in-memory cache of the Registry using the Watcher. If the Registry fails for any 
reason Discovery continues to function.

It can be used to augment the Registry behaviour in a go-micro service and provide a better view 
of the environment. Integration still requires some work.

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
