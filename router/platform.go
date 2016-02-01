package router

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/server"

	proto "github.com/micro/router-srv/proto/router"

	"golang.org/x/net/context"
)

type platform struct {
	opts selector.Options

	client client.Client
	server server.Server

	r proto.RouterClient

	// TODO
	// selector cache service:[]versions
	// stats cache client calls service:[]stats

	sync.RWMutex
	running bool
	exit    chan bool
	stats   map[string]*stats

	cache map[string]*cache
}

var (
	publishInterval = time.Second * 10
)

func newPlatform(opts ...selector.Option) Router {
	options := selector.Options{
		Context: context.TODO(),
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
		opts:   options,
		client: c,
		server: s,
		cache:  make(map[string]*cache),
		stats:  make(map[string]*stats),
		r:      proto.NewRouterClient("go.micro.srv.router", c),
	}
}

func (p *platform) newStats(s *registry.Service, node *registry.Node) {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.stats[node.Id]; ok {
		return
	}

	p.stats[node.Id] = newStats(&registry.Service{
		Name:     s.Name,
		Version:  s.Version,
		Metadata: s.Metadata,
		Nodes:    []*registry.Node{node},
	})
}

func (p *platform) publish() {
	p.RLock()
	defer p.RUnlock()

	opts := p.server.Options()

	// temporarily build client Service
	// should just be pulled from opts.Service()
	// or something like that
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

	// the client service. this is us
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

	// publish all the stats and reset
	for _, stat := range p.stats {
		// create publication
		msg := p.client.NewPublication(StatsTopic, stat.ToProto(service))
		// reset the stats
		stat.Reset()
		// publish message
		go p.client.Publish(context.TODO(), msg)
	}
}

func (p *platform) subscribe() {
	// TODO: subscribe to stream of updates from router
	// send request for every service
	// recv for every service
	// create cache
	// create stats

	return
}

func (p *platform) rselect(service string) (*cache, error) {
	// check the cache
	p.RLock()
	if c, ok := p.cache[service]; ok {
		p.RUnlock()
		return c, nil
	}
	p.RUnlock()

	// not cached

	// call router to get selection for service
	rsp, err := p.r.Select(context.TODO(), &proto.SelectRequest{
		Service: service,
	})

	// error then bail
	if err != nil {
		return nil, err
	}

	// translate from proto to *registry.Service
	var services []*registry.Service
	for _, serv := range rsp.Services {
		rservice := protoToService(serv)
		services = append(services, rservice)

		// create stats
		for _, node := range rservice.Nodes {
			p.newStats(rservice, node)
		}
	}

	// cache the service
	c := newCache(services, rsp.Expires)
	p.Lock()
	p.cache[service] = c
	p.Unlock()

	return c, nil
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

	if c, ok := client.FromContext(options.Context); ok {
		p.client = c
		p.r = proto.NewRouterClient("go.micro.srv.router", c)
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
	p.Lock()
	defer p.Unlock()

	stats, ok := p.stats[node.Id]
	if !ok {
		return
	}

	stats.Record(r, node, d, err)
}

func (p *platform) Stats() ([]*Stats, error) {
	return nil, nil
}

func (p *platform) Select(service string, opts ...selector.SelectOption) (selector.Next, error) {
	// create options
	var options selector.SelectOptions
	for _, o := range opts {
		o(&options)
	}

	// get service from the cache
	// or call the router for list
	cache, err := p.rselect(service)
	if err != nil {
		return nil, err
	}
	return cache.Filter(options.Filters)
}

func (p *platform) Mark(service string, node *registry.Node, err error) {
	p.Lock()
	defer p.Unlock()

	stats, ok := p.stats[node.Id]
	if !ok {
		return
	}

	// mark result for the node
	stats.Mark(service, node, err)
}

func (p *platform) Reset(service string) {
	p.Lock()
	defer p.Unlock()

	// reset stats for the service
	for _, stat := range p.stats {
		if stat.service.Name == service {
			stat.Reset()
		}
	}
}

func (p *platform) Close() error {
	return p.Stop()
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
