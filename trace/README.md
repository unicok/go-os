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
	// Name
	String() string
}

type Span struct {
	Name      string        // Topic / RPC Method
	Id        string        // id of this span
	TraceId   string        // The root trace id
	ParentId  string        // Parent span id
	Timestamp time.Time     // Microseconds from epoch. When span started.
	Duration  time.Duration // Microseconds. Duration of the span.
	Debug     bool          // Should persist no matter what.

	Source      *registry.Service // Originating service
	Destination *registry.Service // Destination service

	sync.Mutex
	Annotations []*Annotation
}

type Annotation struct {
	Timestamp time.Time // Microseconds from epoch
	Type      AnnotationType
	Key       string
	Value     []byte
	Debug     map[string]string
	Service   *registry.Service // Annotator
}

func NewTrace(opts ...Option) Trace {
	return newPlatform(opts...)
}
```

## Supported

- Trace service
- Zipkin
