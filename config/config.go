package config

import (
	"time"
)

// Config is the top level config which aggregates a
// number of sources and provides a single merged
// interface.
type Config interface {
	// Loads config from sources
	Load() error
	// Config values
	Values
	// Combined source
	Source
	// Options for the config
	Options() Options
	// Start/Stop for internal interval updater, etc.
	Start() error
	Stop() error
}

// Source is the source from which config is loaded.
// This may be a file, a url, consul, etc.
type Source interface {
	// Loads ChangeSet from the source
	Read() (*ChangeSet, error)
	// Watch for changes
	Watch() (Watcher, error)
	// Name of source
	String() string
}

// Values loaded within the config
type Values interface {
	// The path could be a nested structure so
	// make it a composable.
	Get(path ...string) Value
}

// Represent a value retrieved from the values loaded
type Value interface {
	Bool() bool
	Int(def int) int
	String(def string) string
	Float64(def float64) float64
	Duration(def string) time.Duration
	StringSlice() []string
	StringMap() map[string]string
	Scan(val interface{}) error
	Bytes() []byte
}

// ChangeSet represents a set an actual source
type ChangeSet struct {
	// The time at which the last change occured
	Timestamp time.Time
	// The raw data set for the change
	Data []byte
	// Hash of the source data
	Checksum string
	// The source of this change
	Source string
}

// The watcher notifies of changes at a granular level.
// Changes() can be called multiple times to retrieve
// new channels. When Stop() is called, all channels
// are closed and the watcher is rendered unusable.
type Watcher interface {
	// Retrieve a watcher on the source
	Changes() <-chan *ChangeSet
	// Stop all channels
	Stop() error
}

/*

Scoped config by labels, environment, datacenter, etc
type Context interface {}

Validate loaded properties
type Validator interface {}

*/

type Option func(o *Options)

type SourceOption func(o *SourceOptions)
