package config

import (
	"time"
)

type Options struct {
	PollInterval time.Duration
	Reader       Reader
	Sources      []Source
}

type SourceOptions struct {
	// Name, Url, etc
	Name string
}

// PollInterval is the time interval at which the sources are polled
// to retrieve config.
func PollInterval(i time.Duration) Option {
	return func(o *Options) {
		o.PollInterval = i
	}
}

// WithSource appends a source to our list of sources.
// This forms a hierarchy whereby all the configs are
// merged down with the last specified as favoured.
func WithSource(s Source) Option {
	return func(o *Options) {
		o.Sources = append(o.Sources, s)
	}
}

// WithReader is the reader used by config to parse
// ChangeSets, merge them and provide values.
// We're not as elegant here in terms of encoding.
func WithReader(r Reader) Option {
	return func(o *Options) {
		o.Reader = r
	}
}

// Source options

// SourceName is an option to provide name of a file,
// a url, key within etcd, consul, zookeeper, etc.
func SourceName(n string) SourceOption {
	return func(o *SourceOptions) {
		o.Name = n
	}
}
