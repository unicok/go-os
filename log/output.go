package log

import (
	"encoding/json"
	"os"
)

type output struct {
	opts OutputOptions

	err error
	f   *os.File
}

func (o *output) Send(e *Event) error {
	if o.f == nil {
		return o.err
	}
	return json.NewEncoder(o.f).Encode(e)
}

func (o *output) Flush() error {
	if o.f == nil {
		return o.err
	}
	return o.f.Sync()
}

func (o *output) Close() error {
	if o.f == nil {
		return o.err
	}
	return o.f.Close()
}

func (o *output) String() string {
	return "json-file"
}

func NewOutput(opts ...OutputOption) Output {
	var options OutputOptions
	for _, o := range opts {
		o(&options)
	}

	if len(options.Name) == 0 {
		options.Name = DefaultOutputName
	}

	f, err := os.OpenFile(options.Name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	return &output{
		opts: options,
		err:  err,
		f:    f,
	}
}
