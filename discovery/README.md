# Discovery - Discovery interface

Provides a high level pluggable abstraction for discovery.

## Interface

While the registry provides service and node information it does not include 
heartbeating or fault tolerant behaviour. Discovery should do both. It should 
cache registry info locally, heartbeat and remove nodes when they stop beating. 
Should also understand where there's massive failure and prevent deleting the 
cache.

```
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
```

## Supported Backends

- Micro registry (any plugins; consul, etcd, memory)
- Platform
