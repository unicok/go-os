package router

import (
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/server"

	"golang.org/x/net/context"
)

type platform struct {
	opts selector.Options
	sync.RWMutex
	running bool
	exit    chan bool

	// fallback selector
	selector selector.Selector

	client client.Client
	server server.Server

	// TODO
	// selector cache service:[]versions
	// stats cache client calls service:[]stats
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
		selector: selector.NewSelector(selector.Registry(options.Registry)),
	}
}

func (p *platform) publish() {
	// TODO: publish stats
	return
}

func (p *platform) subscribe() {
	// TODO: subscribe to stats
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
		p.selector = selector.NewSelector(selector.Registry(options.Registry))
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

func (p *platform) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	// TODO: read from cache
	// fallback to selector
	return p.selector.Select(service, opts...)
}

func (p *platform) Mark(service string, node *registry.Node, err error) {
	// TODO: update stats
	// also update selector
	p.selector.Mark(service, node, err)
}

func (p *platform) Reset(service string) {
	// reset stats?
	// also update selector
	p.selector.Reset(service)
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
