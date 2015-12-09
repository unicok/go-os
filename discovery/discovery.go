package discovery

import (
	"github.com/micro/go-micro/registry"
)

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
