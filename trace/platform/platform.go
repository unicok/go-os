package platform

import (
	"github.com/micro/go-platform/trace"
)

func NewTrace(opts ...trace.Option) trace.Trace {
	return trace.NewTrace(opts...)
}
