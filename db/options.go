package db

import (
	"golang.org/x/net/context"
)

type Options struct {
	Database string
	Table    string

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
