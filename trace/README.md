# Trace - Tracing interface

Provides a pluggable distributed tracing interface

```go
type Trace interface {
	// New span with certain fields preset.
	// Provide parent span if you have it.
	NewSpan(*Span) *Span
	// New context with span
	NewContext(context.Context, *Span) context.Context
	// Return a span from context
	FromContext(context.Context) (*Span, bool)
	// Span to Header
	NewHeader(map[string]string, *Span) map[string]string
	// Get span from header
	FromHeader(map[string]string) (*Span, bool)

	// Collect spans
	Collect(*Span) error
	// Start the collector
	Start() error
	// Stop the collector
	Stop() error
}
```

## Supported

- Platform
- Zipkin
- ?
