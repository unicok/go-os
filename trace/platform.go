package trace

import (
	"github.com/micro/go-micro/context"
)

type platform struct {
	opts Options
}

func newPlatform(opts ...Option) Trace {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}
	return &platform{opts: opt}
}

func (p *platform) Collect(s *Span) error {
	return nil
}

func (p *platform) NewSpan(s *Span) *Span {
	return nil
}

func (p *platform) FromMetadata(md context.Metadata) *Span {
	return nil
}

func (p *platform) ToMetadata(s *Span) context.Metadata {
	return nil
}

func (p *platform) Start() error {
	return nil
}

func (p *platform) Stop() error {
	return nil
}
