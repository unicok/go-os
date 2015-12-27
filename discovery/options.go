package discovery

import (
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
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

func Server(s server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}
