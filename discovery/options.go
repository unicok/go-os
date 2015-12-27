package discovery

import (
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
)

func Client(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

func Interval(i time.Duration) Option {
	return func(o *Options) {
		o.Interval = i
	}
}

func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func UseDiscovery(b bool) Option {
	return func(o *Options) {
		o.Discovery = b
	}
}
