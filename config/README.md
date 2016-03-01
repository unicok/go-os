# Config - Dynamic config interface

Provides a high level pluggable abstraction for dynamic configuration.

## Interface

There's a need for dynamic configuration with namespacing, deltas for rollback, 
watches for changes and an audit log. At a low level we may care about server 
addresses changing, routing information, etc. At a high level there may be a 
need to control business level logic; External API Urls, Pricing information, etc.

```go
// Config is the top level config which aggregates a
// number of sources and provides a single merged
// interface.
type Config interface {
	// Config values
	Values
	// Config options
	Options() Options
	// Start/Stop for internal interval updater, etc.
	Start() error
	Stop() error
	// String name of config; platform
	String() string
}

// Values loaded within the config
type Values interface {
	// The path could be a nested structure so
	// make it a composable.
	// Returns internal cached value
	Get(path ...string) Value
	// Sets internal cached value
	Set(val interface{}, path ...string)
	// Deletes internal cached value
	Del(path ...string)
	// Returns vals as bytes
	Bytes() []byte
}

// Represent a value retrieved from the values loaded
type Value interface {
	Bool(def bool) bool
	Int(def int) int
	String(def string) string
	Float64(def float64) float64
	Duration(def time.Duration) time.Duration
	StringSlice(def []string) []string
	StringMap(def map[string]string) map[string]string
	Scan(val interface{}) error
	Bytes() []byte
}

func NewConfig(opts ...Option) Config {
	return newPlatform(opts...)
}
```

## Supported Backends

- Config service
- File
