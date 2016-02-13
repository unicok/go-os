package trace

import (
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
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
	var okk bool
	var err error

	// Expectation is that we're the initiator of tracing
	// So get trace info from metadata
	md, ok := metadata.FromContext(ctx)
	if !ok {
		// this is a new span
		span = c.t.NewSpan(nil)
	} else {
		// can we gt the span from the header?
		span, okk = c.t.FromHeader(md)
		if !okk {
			// no, ok create one
			span = c.t.NewSpan(nil)
		}
	}

	// if we are the creator
	if !okk {
		// start the span
		span.Annotations = append(span.Annotations, &Annotation{
			Timestamp: time.Now(),
			Type:      AnnStart,
			Service:   c.s,
		})
		// and mark as debug? might want to do this based on a setting
		span.Debug = true
		// set uniq span name
		span.Name = req.Service() + "." + req.Method()
		// set source/dest
		span.Source = c.s
		span.Destination = &registry.Service{Name: req.Service()}
	}

	// set context key
	newCtx := c.t.NewContext(ctx, span)
	// set metadata
	newCtx = metadata.NewContext(newCtx, c.t.NewHeader(md, span))

	// mark client request
	span.Annotations = append(span.Annotations, &Annotation{
		Timestamp: time.Now(),
		Type:      AnnClientRequest,
		Service:   c.s,
	})

	// defer the completion of the span
	defer func() {
		// mark client response
		span.Annotations = append(span.Annotations, &Annotation{
			Timestamp: time.Now(),
			Type:      AnnClientResponse,
			Service:   c.s,
		})

		// if we were the creator
		if !okk {
			var debug map[string]string
			if err != nil {
				debug = map[string]string{"error": err.Error()}
			}
			// mark end of span
			span.Annotations = append(span.Annotations, &Annotation{
				Timestamp: time.Now(),
				Type:      AnnEnd,
				Service:   c.s,
				Debug:     debug,
			})
			span.Duration = time.Now().Sub(span.Timestamp)
		}
		// flush the span to the collector on return
		c.t.Collect(span)
	}()

	// now just make a regular call down the stack
	err = c.Client.Call(newCtx, req, rsp, opts...)
	return err
}

func handlerWrapper(fn server.HandlerFunc, t Trace, s *registry.Service) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		var span *Span
		var err error

		// Expectation is that we're the initiator of tracing
		// So get trace info from metadata
		md, ok := metadata.FromContext(ctx)
		if !ok {
			// this is a new span
			span = t.NewSpan(nil)
			span.Debug = true
		} else {
			// can we gt the span from the header?
			span, ok = t.FromHeader(md)
			if !ok {
				// no, ok create one
				span = t.NewSpan(nil)
			}
			span.Debug = true
		}

		newCtx := t.NewContext(ctx, span)

		// mark client request
		span.Annotations = append(span.Annotations, &Annotation{
			Timestamp: time.Now(),
			Type:      AnnServerRequest,
			Service:   s,
		})

		// defer the completion of the span
		defer func() {
			var debug map[string]string
			if err != nil {
				debug = map[string]string{"error": err.Error()}
			}
			// mark server response
			span.Annotations = append(span.Annotations, &Annotation{
				Timestamp: time.Now(),
				Type:      AnnServerResponse,
				Service:   s,
				Debug:     debug,
			})
			// flush the span to the collector on return
			t.Collect(span)
		}()
		err = fn(newCtx, req, rsp)
		return err
	}
}
