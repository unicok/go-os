package db

import (
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
)

type Options struct {
	Database string
	Table    string

	Client client.Client

	// For alternative options
	Context context.Context
}

func Database(d string) Option {
	return func(o *Options) {
		o.Database = d
	}
}

func Table(t string) Option {
	return func(o *Options) {
		o.Table = t
	}
}

func Client(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}
