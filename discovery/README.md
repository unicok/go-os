# Discovery - Discovery interface

Provides a high level pluggable abstraction for discovery.

## Interface

While the registry provides service and node information it does not include 
heartbeating or fault tolerant behaviour. Discovery should do both. It should 
cache registry info locally, heartbeat and remove nodes when they stop beating. 
Should also understand where there's massive failure and prevent deleting the 
cache.

```
type Discovery interface {
	micro.Registry()
	Run() error // starts heartbeating and caching
	Stop() error // stop heartbeating and clear cache
}

``

## Supported Backends

- micro registry (any plugins; consul, etcd, memory)
