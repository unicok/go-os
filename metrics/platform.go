package metrics

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type platform struct {
	opts Options

	sync.Mutex
	running bool
	buf     chan string
	exit    chan bool
	conn    net.Conn
}

type counter struct {
	id  string
	buf chan string
	f   Fields
}

type gauge struct {
	id  string
	buf chan string
	f   Fields
}

type histogram struct {
	id  string
	buf chan string
	f   Fields
}

var (
	maxBufferSize = 500
)

func newPlatform(opts ...Option) Metrics {
	options := Options{
		Namespace:     DefaultNamespace,
		BatchInterval: DefaultBatchInterval,
		Collectors:    []string{"127.0.0.1:8125"},
		Fields:        make(Fields),
	}

	for _, o := range opts {
		o(&options)
	}

	return &platform{
		opts: options,
		buf:  make(chan string, 1000),
	}
}

func (c *counter) format(v uint64) string {
	keys := []string{c.id}

	for k, v := range c.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|c", strings.Join(keys, ","), v)
}

func (c *counter) Incr(d uint64) {
	c.buf <- c.format(d)
}

func (c *counter) Decr(d uint64) {
	c.buf <- c.format(-d)
}

func (c *counter) Reset() {
	c.buf <- c.format(0)
}

func (c *counter) WithFields(f Fields) Counter {
	nf := make(Fields)

	for k, v := range c.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &counter{
		buf: c.buf,
		id:  c.id,
		f:   nf,
	}
}

func (g *gauge) format(v int64) string {
	keys := []string{g.id}

	for k, v := range g.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|g", strings.Join(keys, ","), v)
}

func (g *gauge) Set(d int64) {
	g.buf <- g.format(d)
}

func (g *gauge) Reset() {
	g.buf <- g.format(0)
}

func (g *gauge) WithFields(f Fields) Gauge {
	nf := make(Fields)

	for k, v := range g.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &gauge{
		buf: g.buf,
		id:  g.id,
		f:   nf,
	}
}

func (h *histogram) format(v int64) string {
	keys := []string{h.id}

	for k, v := range h.f {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("%s:%d|ms", strings.Join(keys, ","), v)
}

func (h *histogram) Record(d int64) {
	h.buf <- h.format(d)
}

func (h *histogram) Reset() {
	h.buf <- h.format(0)
}

func (h *histogram) WithFields(f Fields) Histogram {
	nf := make(Fields)

	for k, v := range h.f {
		nf[k] = v
	}

	for k, v := range f {
		nf[k] = v
	}

	return &histogram{
		buf: h.buf,
		id:  h.id,
		f:   nf,
	}
}

func (p *platform) run() {
	t := time.NewTicker(p.opts.BatchInterval)
	buf := bytes.NewBuffer(nil)

	for {
		select {
		case <-p.exit:
			t.Stop()
			buf.Reset()
			buf = nil
			return
		case v := <-p.buf:
			buf.Write([]byte(fmt.Sprintf("%s.%s\n", p.opts.Namespace, v)))
			if buf.Len() < maxBufferSize {
				continue
			}
			p.conn.Write(buf.Bytes())
			buf.Reset()
		case <-t.C:
			p.conn.Write(buf.Bytes())
			buf.Reset()
		}
	}
}

func (p *platform) Init(opts ...Option) error {
	for _, o := range opts {
		o(&p.opts)
	}
	return nil
}

func (p *platform) Counter(id string) Counter {
	return &counter{
		id:  id,
		buf: p.buf,
		f:   p.opts.Fields,
	}
}

func (p *platform) Gauge(id string) Gauge {
	return &gauge{
		id:  id,
		buf: p.buf,
		f:   p.opts.Fields,
	}
}

func (p *platform) Histogram(id string) Histogram {
	return &histogram{
		id:  id,
		buf: p.buf,
		f:   p.opts.Fields,
	}
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running {
		return nil
	}

	conn, err := net.DialTimeout("udp", p.opts.Collectors[0], time.Second)
	if err != nil {
		return err
	}
	p.conn = conn
	p.exit = make(chan bool)
	p.running = true
	go p.run()
	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()

	if !p.running {
		return nil
	}

	close(p.exit)
	p.exit = nil
	p.running = false
	return p.conn.Close()
}

func (p *platform) String() string {
	return "platform"
}
