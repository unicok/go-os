package trace

import (
	"errors"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/context"
	"github.com/micro/go-micro/registry"
	proto "github.com/micro/go-platform/trace/proto"

	"github.com/pborman/uuid"
	ctx "golang.org/x/net/context"
)

type platform struct {
	opts Options

	spans chan *Span

	sync.Mutex
	running bool
	exit    chan bool
}

func newPlatform(opts ...Option) Trace {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}

	if opt.BatchSize == 0 {
		opt.BatchSize = DefaultBatchSize
	}

	if opt.BatchInterval == time.Duration(0) {
		opt.BatchInterval = DefaultBatchInterval
	}

	if len(opt.Topic) == 0 {
		opt.Topic = TraceTopic
	}

	if opt.Client == nil {
		opt.Client = client.DefaultClient
	}

	return &platform{
		opts:  opt,
		spans: make(chan *Span, 100),
	}
}

func serviceToProto(s *registry.Service) *proto.Service {
	if s == nil || len(s.Nodes) == 0 {
		return nil
	}
	return &proto.Service{
		Name:     s.Name,
		Version:  s.Version,
		Metadata: s.Metadata,
		Nodes: []*proto.Node{&proto.Node{
			Id:       s.Nodes[0].Id,
			Address:  s.Nodes[0].Address,
			Port:     int64(s.Nodes[0].Port),
			Metadata: s.Nodes[0].Metadata,
		}},
	}
}

func toProto(s *Span) *proto.Span {
	var annotations []*proto.Annotation

	for _, a := range s.Annotations {
		annotations = append(annotations, &proto.Annotation{
			Timestamp: a.Timestamp.UnixNano() / 1e3,
			Type:      proto.Annotation_Type(a.Type),
			Key:       a.Key,
			Value:     a.Value,
			Debug:     a.Debug,
			Service:   serviceToProto(a.Service),
		})
	}

	return &proto.Span{
		Name:        s.Name,
		Id:          s.Id,
		TraceId:     s.TraceId,
		ParentId:    s.ParentId,
		Timestamp:   s.Timestamp.UnixNano() / 1e3,
		Duration:    s.Duration.Nanoseconds() / 1e3,
		Debug:       s.Debug,
		Source:      serviceToProto(s.Source),
		Destination: serviceToProto(s.Destination),
		Annotations: annotations,
	}
}

func (p *platform) send(buf []*Span) {
	for _, s := range buf {
		pub := p.opts.Client.NewPublication(p.opts.Topic, toProto(s))
		p.opts.Client.Publish(ctx.TODO(), pub)
	}
}

func (p *platform) run(exit chan bool) {
	t := time.NewTicker(p.opts.BatchInterval)

	var buf []*Span

	for {
		select {
		case s := <-p.spans:
			buf = append(buf, s)
			if len(buf) >= p.opts.BatchSize {
				go p.send(buf)
				buf = nil
			}
		case <-t.C:
			// flush
			if len(buf) > 0 {
				go p.send(buf)
				buf = nil
			}
		case <-exit:
			t.Stop()
			return
		}
	}
}

func (p *platform) Collect(s *Span) error {
	select {
	case p.spans <- s:
	case <-time.After(p.opts.CollectTimeout):
		return errors.New("Timed out sending span")
	}
	return nil
}

func (p *platform) NewSpan(s *Span) *Span {
	if s == nil {
		return &Span{
			Id:        uuid.NewUUID().String(),
			TraceId:   uuid.NewUUID().String(),
			ParentId:  "0",
			Timestamp: time.Now(),
			Source:    p.opts.Service,
		}
	}

	if len(s.TraceId) == 0 {
		s.TraceId = uuid.NewUUID().String()
	}
	if len(s.ParentId) == 0 {
		s.ParentId = "0"
	}
	if len(s.Id) == 0 {
		s.Id = uuid.NewUUID().String()
	}

	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	return &Span{
		Id:        s.Id,
		TraceId:   s.TraceId,
		ParentId:  s.ParentId,
		Timestamp: s.Timestamp,
	}
}

func (p *platform) FromMetadata(md context.Metadata) *Span {
	var debug bool
	if md[DebugHeader] == "1" {
		debug = true
	}

	return p.NewSpan(&Span{
		Id:       md[SpanHeader],
		TraceId:  md[TraceHeader],
		ParentId: md[ParentHeader],
		Debug:    debug,
	})
}

func (p *platform) ToMetadata(s *Span) context.Metadata {
	debug := "0"
	if s.Debug {
		debug = "1"
	}

	return context.Metadata{
		SpanHeader:   s.Id,
		TraceHeader:  s.TraceId,
		ParentHeader: s.ParentId,
		DebugHeader:  debug,
	}
	return nil
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running {
		return nil
	}

	ch := make(chan bool)
	go p.run(ch)
	p.exit = ch
	p.running = true
	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()

	if !p.running {
		return nil
	}

	close(p.exit)
	p.running = false
	p.exit = nil
	return nil
}
