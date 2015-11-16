# Metrics - Instrumentation interface
 
Provides a high level abstraction to instrument metrics.

## Interface

```
type Metrics interface {
	Gauge(rate float64, bucket string, n ...int)
	Counter(rate float64, bucket string, n ...int)
	Timing(rate float64, bucket string, d ...time.Duration)
}
```

## Supported Log stores

- Graphite
- InfluxDB
- ?
