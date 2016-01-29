package router

import (
	"github.com/micro/go-micro/selector"
)

// The router is the client interface to a
// global service loadbalancer (GSLB).
// Metrics are batched and published to
// a router which has a view of the whole
// system.
type Router interface {
	selector.Selector
	Start() error
	Stop() error
}

func NewRouter(opts ...selector.Option) Router {
	return newPlatform(opts...)
}
