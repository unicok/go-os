# Monitor

```go
// The monitor aggregates, observes and publishes
// information about the current process.
// This includes status; started/running/stopped,
// stats; cpu, memory, runtime and healthchecks.
type Monitor interface {
	Checker
	Stats
	Start() error
	Stop() error
	String() string
}

// Checker interface allows creation of healthchecks
type Checker interface {
	NewHealthChecker(id, desc string, fn HealthCheck) HealthChecker
	Register(HealthChecker) error
	Deregister(HealthChecker) error
	HealthChecks() ([]HealthChecker, error)
}

// represents a healthcheck function
type HealthChecker interface {
	// Unique id of the healthcheck
	Id() string
	// Description of what the healthcheck does
	Description() string
	// Runs the the healthcheck
	Run() (map[string]string, error)
	// Returns the status of the last run
	// Better where the healthcheck is expensive to run
	Status() (map[string]string, error)
}

// stats interface allows recording of endpoint stats
type Stats interface {
	RecordStat(r Request, d time.Duration, err error)
}

type HealthCheck func() (map[string]string, error)

func NewMonitor(opts ...Option) Monitor {
	return newPlatform(opts...)
}
```

## Supported Backends

- Monitor service
