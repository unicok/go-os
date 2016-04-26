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
	exit chan bool
	opts selector.Options

	client client.Client
	server server.Server

	r proto.RouterClient

	// TODO
	// selector cache service:[]versions
	// stats cache client calls service:[]stats

	sync.RWMutex
	stats map[string]*stats
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

	p := &platform{
		exit:   make(chan bool),
		opts:   options,
		client: c,
		server: s,
		cache:  make(map[string]*cache),
		stats:  make(map[string]*stats),
		r:      proto.NewRouterClient("go.micro.srv.router", c),
	}

	go p.run()
	return p
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

	streams := make(map[string]bool)
	t := time.NewTicker(publishInterval)

	fn := func(service string, exit chan bool) {
		for {
			select {
			case <-exit:
				return
			default:
				p.stream(service)
				time.Sleep(time.Millisecond * 100)
			}
		}
	}

	for {
		select {
		case <-p.exit:
			t.Stop()
			return
		case <-t.C:
			p.RLock()
			cache := p.cache
			p.RUnlock()

			for name, _ := range cache {
				if _, ok := streams[name]; ok {
					continue
				}
				fn(name, p.exit)
				streams[name] = true
			}
		}
	}
}

func (p *platform) stream(service string) {
	stream, err := p.r.SelectStream(context.TODO(), &proto.SelectRequest{Service: service})
	if err != nil {
		return
	}

	exit := make(chan bool)

	defer func() {
		close(exit)
	}()

	go func() {
		select {
		case <-exit:
			// probably errored
		case <-p.exit:
			stream.Close()
		}
	}()

	for {
		rsp, err := stream.Recv()
		if err != nil {
			return
		}

		var services []*registry.Service
		nodes := make(map[string]bool)

		for _, serv := range rsp.Services {
			rservice := protoToService(serv)
			services = append(services, rservice)

			// create stats
			for _, node := range rservice.Nodes {
				p.newStats(rservice, node)
				nodes[node.Id] = true
			}
		}

		p.Lock()
		// delete nodes from stats that have been removed
		// we might actually lost stats by doing this
		// TODO: move to a reaper
		if s, ok := p.cache[service]; ok {
			for node, _ := range s.nodes {
				if _, ok := nodes[node]; !ok {
					delete(p.stats, node)
				}
			}
		}

		// cache the service
		p.cache[service] = newCache(services, rsp.Expires)
		p.Unlock()
	}
}

func (p *platform) rselect(service string) (*cache, error) {
	// check the cache
	p.RLock()
	if c, ok := p.cache[service]; ok {
		if c.expires == -1 || c.expires > time.Now().Unix() {
			p.RUnlock()
			return c, nil
		}
	}
	p.RUnlock()

	// not cached or expired

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

	go p.subscribe()

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

func (p *platform) Close() error {
	select {
	case <-p.exit:
		return nil
	default:
		close(p.exit)
	}
	return nil
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

func (p *platform) String() string {
	return "platform"
}
