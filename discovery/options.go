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

// Service specifies whether to use the platform Discovery service.
// Because go-os/discovery can be used as just a read layer cache
// we can disable the discovery service itself. It's on by default.
func Service(b bool) Option {
	return func(o *Options) {
		o.Discovery = b
	}
}
