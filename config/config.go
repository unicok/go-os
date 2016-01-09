package config

import (
	"time"
)

// Config is the top level config which aggregates a
// number of sources and provides a single merged
// interface.
type Config interface {
	// Load the values from sources and parse
	Load() error
	// Config values
	Values
	// Config options
	Options() Options
	// Start/Stop for internal interval updater, etc.
	Start() error
	Stop() error
	// String name of config; platform
	String()
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

// Source is the source from which config is loaded.
// This may be a file, a url, consul, etc.
type Source interface {
	// Loads ChangeSet from the source
	Read() (*ChangeSet, error)
	// Name of source
	String() string
}

// Parser takes a changeset from a source and returns Values.
// E.g reads ChangeSet as JSON and can merge down
type Parser interface {
	// Parse ChangeSets
	Parse(...*ChangeSet) (Values, error)
	// Name of parser; json
	String() string
}

// ChangeSet represents a set an actual source
type ChangeSet struct {
	// The time at which the last change occured
	Timestamp time.Time
	// The raw data set for the change
	Data []byte
	// Hash of the source data
	Checksum string
	// The source of this change; file, consul, etcd
	Source string
}

type Option func(o *Options)

type SourceOption func(o *SourceOptions)
