package router

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/blacklist"
	"github.com/micro/go-micro/server"

	"github.com/micro/go-platform/router/proto"

	"golang.org/x/net/context"
)

type platform struct {
	opts selector.Options

	// fallback selector
	selector selector.Selector

	client client.Client
	server server.Server

	// TODO
	// selector cache service:[]versions
	// stats cache client calls service:[]stats

	sync.RWMutex
	running bool
	exit    chan bool
	stats   map[string]*stats
}

var (
	publishInterval = time.Second * 10
)

func newPlatform(opts ...selector.Option) Router {
	options := selector.Options{
		Registry: registry.DefaultRegistry,
		Context:  context.TODO(),
	}

	for _, o := range opts {
		o(&options)
	}

	c, ok := client.FromContext(options.Context)
	if !ok {
		c = client.DefaultClient
	}

	s, ok := server.FromContext(options.Context)
	if !ok {
		s = server.DefaultServer
	}

	return &platform{
		opts:     options,
		client:   c,
		server:   s,
		selector: blacklist.NewSelector(selector.Registry(options.Registry)),
		stats:    make(map[string]*stats),
	}
}

func serviceToProto(s *registry.Service) *router.Service {
	if s == nil || len(s.Nodes) == 0 {
		return nil
	}
	return &router.Service{
		Name:     s.Name,
		Version:  s.Version,
		Metadata: s.Metadata,
		Nodes: []*router.Node{&router.Node{
			Id:       s.Nodes[0].Id,
			Address:  s.Nodes[0].Address,
			Port:     int64(s.Nodes[0].Port),
			Metadata: s.Nodes[0].Metadata,
		}},
	}
}

func (p *platform) publish() {
	p.RLock()
	defer p.RUnlock()

	opts := p.server.Options()

	var addr, host string
	var port int

	if len(opts.Advertise) > 0 {
		addr = opts.Advertise
	} else {
		addr = opts.Address
	}

	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		i, _ := strconv.ParseInt(parts[1], 10, 64)
		host = parts[0]
		port = int(i)
	} else {
		host = addr
	}

	service := &registry.Service{
		Name:    opts.Name,
		Version: opts.Version,
		Nodes: []*registry.Node{&registry.Node{
			Id:       opts.Id,
			Address:  host,
			Port:     port,
			Metadata: opts.Metadata,
		}},
	}

	// publish all the stats
	for _, stat := range p.stats {
		pub := p.client.NewPublication(StatsTopic, stat.ToProto(service))
		go p.client.Publish(context.TODO(), pub)
	}
}

func (p *platform) subscribe() {
	// TODO: subscribe to stream of updates from router
	return
}

func (p *platform) run() {
	t := time.NewTicker(publishInterval)

	for {
		select {
		case <-p.exit:
			t.Stop()
			return
		case <-t.C:
			p.publish()
		}
	}
}

func (p *platform) Init(opts ...selector.Option) error {
	var options selector.Options
	for _, o := range opts {
		o(&options)
	}

	// TODO: Fix. This might all be really bad and hacky

	if options.Registry != nil {
		p.opts.Registry = options.Registry
		p.selector = blacklist.NewSelector(selector.Registry(options.Registry))
	}

	if c, ok := client.FromContext(options.Context); ok {
		p.client = c
	}

	if s, ok := server.FromContext(options.Context); ok {
		p.server = s
	}

	return nil
}

func (p *platform) Options() selector.Options {
	return p.opts
}

func (p *platform) Record(r Request, node *registry.Node, d time.Duration, err error) {
	return
}

func (p *platform) Stats() ([]*Stats, error) {
	return nil, nil
}

func (p *platform) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	// TODO: read from cache
	// fallback to selector
	return p.selector.Select(service, opts...)
}

func (p *platform) Mark(service string, node *registry.Node, err error) {
	p.Lock()
	defer p.Unlock()

	p.selector.Mark(service, node, err)

	stats, ok := p.stats[node.Id]
	if !ok {
		return
	}
	stats.Mark(service, node, err)
}

func (p *platform) Reset(service string) {
	p.Lock()
	defer p.Unlock()

	p.selector.Reset(service)

	// reset counters and stats
	for _, stat := range p.stats {
		if stat.service.Name != service {
			continue
		}
		stat.Reset()
	}
}

func (p *platform) Close() error {
	// Should we clear the cache?
	// Should we not stop?
	p.Stop()
	// close the selector
	return p.selector.Close()
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running {
		return nil
	}

	p.running = true
	p.exit = make(chan bool)
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
	return nil
}

func (p *platform) String() string {
	return "platform"
}
