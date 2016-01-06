package kv

import (
	"golang.org/x/net/context"
)

type Options struct {
	Servers []string

	// Alternative options set using Context
	Context context.Context
}

func Servers(s []string) Option {
	return func(o *Options) {
		o.Servers = s
	}
}
