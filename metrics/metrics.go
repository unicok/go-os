package metrics

// Metrics provides a way to instrument application data
type Metrics interface {
	Counter(id string) Counter
	Gauge(id string) Gauge
	Histogram(id string) Histogram
	// Start/Stop a batched collector
	Start() error
	Stop() error
}

type Counter interface {
	// Increment by the given value
	Incr(d uint64)
	// Decrement by the given value
	Decr(d uint64)
	// Reset the counter
	Reset()
}

type Gauge interface {
	// Set the gauge value
	Set(d int64)
}

type Histogram interface {
	// Record a timing
	Record(d int64)
}

type Option func(o *Options)
