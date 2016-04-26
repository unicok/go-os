package log

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
)

type platform struct {
	*logger
	opts Options
}

type logger struct {
	f  Fields
	fn logFunc
}

type logFunc func(l Level, f Fields, m string) error

func newPlatform(opts ...Option) Log {
	options := Options{
		Level:   DefaultLevel,
		Fields:  make(Fields),
		Context: context.TODO(),
	}

	for _, o := range opts {
		o(&options)
	}

	if len(options.Outputs) == 0 {
		options.Outputs = []Output{NewOutput()}
	}

	p := &platform{
		&logger{
			f: make(Fields),
		},
		options,
	}

	// so ugly
	p.logger.fn = p.log

	return p
}

func (p *platform) log(l Level, f Fields, m string) error {
	// discard if we're not at the right level
	if l < p.opts.Level {
		return nil
	}

	fields := make(Fields)

	for k, v := range p.opts.Fields {
		fields[k] = v
	}

	for k, v := range f {
		fields[k] = v
	}

	e := &Event{
		Level:     l,
		Fields:    fields,
		Timestamp: time.Now().UnixNano(),
		Message:   m,
	}

	var gerr error
	for _, o := range p.opts.Outputs {
		if err := o.Send(e); err != nil {
			gerr = err
		}
	}
	return gerr
}

func (p *platform) Close() error {
	return nil
}

func (p *platform) Init(opts ...Option) error {
	for _, o := range opts {
		o(&p.opts)
	}
	return nil
}

func (p *platform) Options() Options {
	return p.opts
}

func (p *platform) String() string {
	return "platform"
}

func (l *logger) Debug(args ...interface{}) {
	l.fn(DebugLevel, l.f, fmt.Sprint(args...))
}

func (l *logger) Info(args ...interface{}) {
	l.fn(InfoLevel, l.f, fmt.Sprint(args...))
}

func (l *logger) Error(args ...interface{}) {
	l.fn(ErrorLevel, l.f, fmt.Sprint(args...))
}

func (l *logger) Fatal(args ...interface{}) {
	l.fn(FatalLevel, l.f, fmt.Sprint(args...))
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.fn(DebugLevel, l.f, fmt.Sprintf(format, args...))
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.fn(InfoLevel, l.f, fmt.Sprintf(format, args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.fn(ErrorLevel, l.f, fmt.Sprintf(format, args...))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.fn(FatalLevel, l.f, fmt.Sprintf(format, args...))
}

func (l *logger) Log(level Level, args ...interface{}) {
	l.fn(level, l.f, fmt.Sprint(args...))
}

func (l *logger) Logf(level Level, format string, args ...interface{}) {
	l.fn(level, l.f, fmt.Sprintf(format, args...))
}

func (l *logger) WithFields(f Fields) Logger {
	fields := make(Fields)

	for k, v := range l.f {
		fields[k] = v
	}

	for k, v := range f {
		fields[k] = v
	}

	return &logger{
		f:  fields,
		fn: l.fn,
	}
}
