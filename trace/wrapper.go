package trace

import (
	"github.com/micro/go-micro/client"
	co "github.com/micro/go-micro/context"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"
	"time"

	"golang.org/x/net/context"
)

type clientWrapper struct {
	client.Client
	t Trace
	s *registry.Service
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	var span *Span
	md, ok := co.GetMetadata(ctx)
	if !ok {
		span = c.t.NewSpan(nil)
	} else {
		span = c.t.FromMetadata(md)
	}

	span.Debug = true

	newCtx := co.WithMetadata(ctx, c.t.ToMetadata(span))
	// request

	span.Annotations = append(span.Annotations, &Annotation{
		Timestamp: time.Now(),
		Type:      AnnClientRequest,
		Service:   c.s,
	})
	// response
	defer func() {
		span.Annotations = append(span.Annotations, &Annotation{
			Timestamp: time.Now(),
			Type:      AnnClientResponse,
			Service:   c.s,
		})
		c.t.Collect(span)
	}()
	return c.Client.Call(newCtx, req, rsp, opts...)
}

func handlerWrapper(fn server.HandlerFunc, t Trace, s *registry.Service) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		var span *Span
		md, ok := co.GetMetadata(ctx)
		if !ok {
			span = t.NewSpan(nil)
		} else {
			span = t.FromMetadata(md)
		}

		newCtx := co.WithMetadata(ctx, t.ToMetadata(span))
		// request
		span.Annotations = append(span.Annotations, &Annotation{
			Timestamp: time.Now(),
			Type:      AnnServerRequest,
			Service:   s,
		})
		// response
		defer func() {
			span.Annotations = append(span.Annotations, &Annotation{
				Timestamp: time.Now(),
				Type:      AnnClientResponse,
				Service:   s,
			})
			t.Collect(span)
		}()
		return fn(newCtx, req, rsp)
	}
}
