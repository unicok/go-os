package logrus

/*
	logrus output integrated logrus output.
*/

import (
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/micro/go-os/log"
)

type output struct {
	*logrus.Logger

	opts log.OutputOptions
	err  error
}

func (o *output) Send(e *log.Event) error {
	if o.Logger == nil {
		return o.err
	}

	fields := logrus.Fields{}
	for k, v := range e.Fields {
		fields[k] = v
	}

	entry := o.WithFields(fields)
	switch e.Level {
	case log.DebugLevel:
		entry.Debug(e.Message)
	case log.InfoLevel:
		entry.Info(e.Message)
	case log.WarnLevel:
		entry.Warn(e.Message)
	case log.ErrorLevel:
		entry.Error(e.Message)
	case log.FatalLevel:
		entry.Fatal(e.Message)
	}

	return nil
}

func (o *output) Flush() error {
	if o.Logger == nil {
		return o.err
	}
	return o.Flush()
}

func (o *output) Close() error {
	if o.Logger == nil {
		return o.err
	}
	return o.Close()
}

func (o *output) String() string {
	return "logrus"
}

func NewOutput(opts ...log.OutputOption) log.Output {
	var options log.OutputOptions
	for _, o := range opts {
		o(&options)
	}

	if len(options.Name) == 0 {
		options.Name = log.DefaultOutputName
	}

	l := logrus.New()
	//Log level is decided by the log of go-os
	l.Level = logrus.PanicLevel

	return &output{
		opts:   options,
		err:    errors.New("logrus initialize failed"),
		Logger: l,
	}
}
