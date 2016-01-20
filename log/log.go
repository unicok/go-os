package log

const (
	DebugLevel Level = 0
	InfoLevel  Level = 1
	WarnLevel  Level = 2
	ErrorLevel Level = 3
	FatalLevel Level = 4
)

type Level int32

type Fields map[string]string

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
	Debuf(format string, args ...interface{})
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
	Level     Level
	Fields    Fields
	Timestamp int64
	Message   string
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

type Option func(o *Options)
