package discovery

import (
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/registry"
)

const (
	heartBeatTopic = "micro.discovery.heartbeat"
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

type Options struct {
	Registry registry.Registry
	Broker   broker.Broker
}

type Option func(*Options)

func NewDiscovery(opts ...Option) Discovery {
	return newPlatform(opts...)
}
