# Log - Logging interface
 
Provides a high level abstraction for centralised logging.

## Interface

Potentially offload this to [logrus](https://github.com/Sirupsen/logrus)

```go
// A structure log interface which can
// output to multiple backends.
type Log interface {
	Init(opts ...Option) error
	Options() Options
	// the logging interface
	Logger
	// We could be flushing logs on an interval basis
	// Or sending specific stats to the log service
	// Or receive events about changing log config
	Start() error
	Stop() error
	// Name
	String() string
}

type Logger interface {
	// Logger interface
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	// Formatted logger
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	// Specify your own levels
	Log(l Level, args ...interface{})
	Logf(l Level, format string, args ...interface{})
	// Returns with extra fields
	WithFields(f Fields) Logger
}

// Event represents a single log event
type Event struct {
	Level     Level  `json:"level"`
	Fields    Fields `json:"fields"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

// An output represents a file, indexer, syslog, etc
type Output interface {
	// Send an event
	Send(*Event) error
	// Flush any buffered events
	Flush() error
	// Discard the output
	Close() error
	// Name of output
	String() string
}

func NewLog(opts ...Option) Log {
	return newPlatform(opts...)
}
```

## Supported Backends

- File
