package zipkin

import (
	"encoding/binary"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/micro/go-micro/context"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-platform/trace"
	"github.com/micro/go-platform/trace/zipkin/thrift/gen-go/zipkincore"

	"github.com/Shopify/sarama"
	"github.com/apache/thrift/lib/go/thrift"
)

type zipkin struct {
	opts  trace.Options
	spans chan *trace.Span

	sync.Mutex
	running bool
	exit    chan bool
}

var (
	TraceTopic = "zipkin"

	TraceHeader  = "X-B3-TraceId"
	SpanHeader   = "X-B3-SpanId"
	ParentHeader = "X-B3-ParentSpanId"
	SampleHeader = "X-B3-Sampled"
)

func random() int64 {
	return rand.Int63() & 0x001fffffffffffff
}

func newZipkin(opts ...trace.Option) trace.Trace {
	var opt trace.Options
	for _, o := range opts {
		o(&opt)
	}

	if opt.BatchSize == 0 {
		opt.BatchSize = trace.DefaultBatchSize
	}

	if opt.BatchInterval == time.Duration(0) {
		opt.BatchInterval = trace.DefaultBatchInterval
	}

	if len(opt.Topic) == 0 {
		opt.Topic = TraceTopic
	}

	return &zipkin{
		opts:  opt,
		spans: make(chan *trace.Span, 100),
	}
}

func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func toEndpoint(s *registry.Service) *zipkincore.Endpoint {
	if s == nil || len(s.Nodes) == 0 {
		return nil
	}

	addrs, err := net.LookupIP(s.Nodes[0].Address)
	if err != nil {
		return nil
	}
	if len(addrs) == 0 {
		return nil
	}
	ep := zipkincore.NewEndpoint()
	binary.LittleEndian.PutUint32(addrs[0], (uint32)(ep.Ipv4))
	ep.Port = int16(s.Nodes[0].Port)
	ep.ServiceName = s.Name
	return ep
}

func toThrift(s *trace.Span) *zipkincore.Span {
	span := &zipkincore.Span{
		TraceID:   toInt64(s.TraceId),
		Name:      s.Target,
		ID:        toInt64(s.Id),
		ParentID:  thrift.Int64Ptr(toInt64(s.ParentId)),
		Debug:     s.Debug,
		Timestamp: thrift.Int64Ptr(s.Timestamp.UnixNano() / 1e3),
		Duration:  thrift.Int64Ptr(s.Duration.Nanoseconds() / 1e3),
	}

	for _, a := range s.Annotations {
		if len(a.Value) > 0 || a.Debug != nil {
			span.BinaryAnnotations = append(span.BinaryAnnotations, &zipkincore.BinaryAnnotation{
				Key:            a.Key,
				Value:          a.Value,
				AnnotationType: zipkincore.AnnotationType_BYTES,
				Host:           toEndpoint(a.Service),
			})
		} else {
			span.Annotations = append(span.Annotations, &zipkincore.Annotation{
				Timestamp: a.Timestamp.UnixNano() / 1e3,
				Value:     a.Key,
				Host:      toEndpoint(a.Service),
			})
		}
	}

	return span
}

func (z *zipkin) pub(s *zipkincore.Span, pr sarama.SyncProducer) {
	t := thrift.NewTMemoryBufferLen(1024)
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Write(p); err != nil {
		return
	}

	m := &sarama.ProducerMessage{
		Topic: z.opts.Topic,
		Value: sarama.ByteEncoder(t.Buffer.Bytes()),
	}
	pr.SendMessage(m)
}

func (z *zipkin) run(ch chan bool, p sarama.SyncProducer) {
	t := time.NewTicker(z.opts.BatchInterval)

	var buf []*trace.Span

	for {
		select {
		case s := <-z.spans:
			buf = append(buf, s)
			if len(buf) >= z.opts.BatchSize {
				go z.send(buf, p)
				buf = nil
			}
		case <-t.C:
			// flush
			if len(buf) > 0 {
				go z.send(buf, p)
				buf = nil
			}
		case <-ch:
			// exit
			t.Stop()
			p.Close()
			return
		}
	}
}

func (z *zipkin) send(b []*trace.Span, p sarama.SyncProducer) {
	for _, span := range b {
		z.pub(toThrift(span), p)
	}
}

func (z *zipkin) Collect(s *trace.Span) error {
	z.spans <- s
	return nil
}

func (z *zipkin) NewSpan(s *trace.Span) *trace.Span {
	if s == nil {
		return &trace.Span{
			Id:        strconv.FormatInt(random(), 10),
			TraceId:   strconv.FormatInt(random(), 10),
			ParentId:  "0",
			Timestamp: time.Now(),
			Source:    z.opts.Service,
		}
	}

	if _, err := strconv.ParseInt(s.TraceId, 16, 64); err != nil {
		s.TraceId = strconv.FormatInt(random(), 10)
	}
	if _, err := strconv.ParseInt(s.ParentId, 16, 64); err != nil {
		s.ParentId = "0"
	}
	if _, err := strconv.ParseInt(s.Id, 16, 64); err != nil {
		s.Id = strconv.FormatInt(random(), 10)
	}

	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	return &trace.Span{
		Id:        s.Id,
		TraceId:   s.TraceId,
		ParentId:  s.ParentId,
		Timestamp: s.Timestamp,
	}
}

func (z *zipkin) FromMetadata(md context.Metadata) *trace.Span {
	var debug bool
	if md[SampleHeader] == "1" {
		debug = true
	}

	return z.NewSpan(&trace.Span{
		Id:       md[SpanHeader],
		TraceId:  md[TraceHeader],
		ParentId: md[ParentHeader],
		Debug:    debug,
	})
}

func (z *zipkin) ToMetadata(s *trace.Span) context.Metadata {
	sample := "0"
	if s.Debug {
		sample = "1"
	}

	return context.Metadata{
		SpanHeader:   s.Id,
		TraceHeader:  s.TraceId,
		ParentHeader: s.ParentId,
		SampleHeader: sample,
	}
}

func (z *zipkin) Start() error {
	z.Lock()
	defer z.Unlock()

	if z.running {
		return nil
	}

	p, err := sarama.NewSyncProducer(z.opts.Collectors, sarama.NewConfig())
	if err != nil {
		return err
	}

	ch := make(chan bool)
	go z.run(ch, p)
	z.exit = ch
	z.running = true
	return nil
}

func (z *zipkin) Stop() error {
	z.Lock()
	defer z.Unlock()

	if !z.running {
		return nil
	}

	close(z.exit)
	z.running = false
	z.exit = nil
	return nil
}

func NewTrace(opts ...trace.Option) trace.Trace {
	return newZipkin(opts...)
}
