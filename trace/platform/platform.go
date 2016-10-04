package platform

import (
	"github.com/micro/go-os/trace"
)

func NewTrace(opts ...trace.Option) trace.Trace {
	return trace.NewTrace(opts...)
}
