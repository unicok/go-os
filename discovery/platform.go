package discovery

import (
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"

	proto2 "github.com/micro/discovery-srv/proto/registry"
	proto "github.com/micro/go-platform/discovery/proto"
	"golang.org/x/net/context"
)

type platform struct {
	opts    Options
	exit    chan bool
	watcher registry.Watcher

	reg proto2.RegistryClient

	sync.RWMutex
	heartbeats map[string]*proto.Heartbeat
	cache      map[string][]*registry.Service
}

type watcher struct {
	wc proto2.Registry_WatchClient
}

func newPlatform(opts ...Option) Discovery {
	opt := Options{
		Discovery: true,
	}

	for _, o := range opts {
		o(&opt)
	}

	if opt.Registry == nil {
		opt.Registry = registry.DefaultRegistry
	}

	if opt.Client == nil {
		opt.Client = client.DefaultClient
	}

	if opt.Interval == time.Duration(0) {
		opt.Interval = time.Second * 30
	}

	return &platform{
		opts:       opt,
		heartbeats: make(map[string]*proto.Heartbeat),
		cache:      make(map[string][]*registry.Service),
		reg:        proto2.NewRegistryClient("go.micro.srv.discovery", opt.Client),
	}
}

func values(v []*registry.Value) []*proto.Value {
	if len(v) == 0 {
		return []*proto.Value{}
	}

	var vs []*proto.Value
	for _, vi := range v {
		vs = append(vs, &proto.Value{
			Name:   vi.Name,
			Type:   vi.Type,
			Values: values(vi.Values),
		})
	}
	return vs
}

func toValues(v []*proto.Value) []*registry.Value {
	if len(v) == 0 {
		return []*registry.Value{}
	}

	var vs []*registry.Value
	for _, vi := range v {
		vs = append(vs, &registry.Value{
			Name:   vi.Name,
			Type:   vi.Type,
			Values: toValues(vi.Values),
		})
	}
	return vs
}

func toProto(s *registry.Service) *proto.Service {
	var endpoints []*proto.Endpoint
	for _, ep := range s.Endpoints {
		var request, response *proto.Value

		if ep.Request != nil {
			request = &proto.Value{
				Name:   ep.Request.Name,
				Type:   ep.Request.Type,
				Values: values(ep.Request.Values),
			}
		}

		if ep.Response != nil {
			response = &proto.Value{
				Name:   ep.Response.Name,
				Type:   ep.Response.Type,
				Values: values(ep.Response.Values),
			}
		}

		endpoints = append(endpoints, &proto.Endpoint{
			Name:     ep.Name,
			Request:  request,
			Response: response,
			Metadata: ep.Metadata,
		})
	}

	var nodes []*proto.Node

	for _, node := range s.Nodes {
		nodes = append(nodes, &proto.Node{
			Id:       node.Id,
			Address:  node.Address,
			Port:     int64(node.Port),
			Metadata: node.Metadata,
		})
	}

	return &proto.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: endpoints,
		Nodes:     nodes,
	}
}

func toService(s *proto.Service) *registry.Service {
	var endpoints []*registry.Endpoint
	for _, ep := range s.Endpoints {
		var request, response *registry.Value

		if ep.Request != nil {
			request = &registry.Value{
				Name:   ep.Request.Name,
				Type:   ep.Request.Type,
				Values: toValues(ep.Request.Values),
			}
		}

		if ep.Response != nil {
			response = &registry.Value{
				Name:   ep.Response.Name,
				Type:   ep.Response.Type,
				Values: toValues(ep.Response.Values),
			}
		}

		endpoints = append(endpoints, &registry.Endpoint{
			Name:     ep.Name,
			Request:  request,
			Response: response,
			Metadata: ep.Metadata,
		})
	}

	var nodes []*registry.Node
	for _, node := range s.Nodes {
		nodes = append(nodes, &registry.Node{
			Id:       node.Id,
			Address:  node.Address,
			Port:     int(node.Port),
			Metadata: node.Metadata,
		})
	}

	return &registry.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: endpoints,
		Nodes:     nodes,
	}
}

func (w *watcher) Next() (*registry.Result, error) {
	r, err := w.wc.RecvR()
	if err != nil {
		return nil, err
	}

	return &registry.Result{
		Action:  r.Result.Action,
		Service: toService(r.Result.Service),
	}, nil
}

func (w *watcher) Stop() {
	w.wc.Close()
}

func (p *platform) heartbeat(t *time.Ticker) {
	for _ = range t.C {
		p.RLock()
		for _, hb := range p.heartbeats {
			hb.Timestamp = time.Now().Unix()
			pub := p.opts.Client.NewPublication(HeartbeatTopic, hb)
			p.opts.Client.Publish(context.TODO(), pub)
		}
		p.RUnlock()
	}
}

func (p *platform) watch(ch chan *registry.Result) {
	p.RLock()
	watch := p.watcher
	p.RUnlock()

	for {
		next, err := watch.Next()
		if err != nil {
			w, err := p.Watch()
			if err != nil {
				return
			}
			p.Lock()
			p.watcher = w
			p.Unlock()
			watch = w
			continue
		}
		ch <- next
	}
}

func (p *platform) run() {
	ch := make(chan *registry.Result)
	t := time.NewTicker(p.opts.Interval)

	go p.watch(ch)
	go p.heartbeat(t)

	for {
		select {
		case <-p.exit:
			t.Stop()
			return
		case next, ok := <-ch:
			if !ok {
				return
			}
			p.update(next)
		}
	}
}

func (p *platform) update(res *registry.Result) {
	p.Lock()
	defer p.Unlock()

	if res == nil || res.Service == nil {
		return
	}

	services, ok := p.cache[res.Service.Name]
	if !ok {
		// no service found, let's not go through a convoluted process
		if (res.Action == "create" || res.Action == "update") && len(res.Service.Version) != 0 {
			p.cache[res.Service.Name] = []*registry.Service{res.Service}
		}
		return
	}

	if len(res.Service.Nodes) == 0 {
		switch res.Action {
		case "delete":
			delete(p.cache, res.Service.Name)
		}
		return
	}

	// existing service found
	var service *registry.Service
	var index int
	for i, s := range services {
		if s.Version == res.Service.Version {
			service = s
			index = i
		}
	}

	switch res.Action {
	case "create", "update":
		if service == nil {
			services = append(services, res.Service)
			p.cache[res.Service.Name] = services
			return
		}

		// append old nodes to new service
		for _, cur := range service.Nodes {
			var seen bool
			for _, node := range res.Service.Nodes {
				if cur.Id == node.Id {
					seen = true
					break
				}
			}
			if !seen {
				res.Service.Nodes = append(res.Service.Nodes, cur)
			}
		}

		services[index] = res.Service
		p.cache[res.Service.Name] = services
	case "delete":
		if service == nil {
			return
		}

		var nodes []*registry.Node

		// filter cur nodes to remove the dead one
		for _, cur := range service.Nodes {
			var seen bool
			for _, del := range res.Service.Nodes {
				if del.Id == cur.Id {
					seen = true
					break
				}
			}
			if !seen {
				nodes = append(nodes, cur)
			}
		}

		if len(nodes) == 0 {
			if len(services) == 1 {
				delete(p.cache, service.Name)
			} else {
				var srvs []*registry.Service
				for _, s := range services {
					if s.Version != service.Version {
						srvs = append(srvs, s)
					}
				}
				p.cache[service.Name] = srvs
			}
			return
		}

		service.Nodes = nodes
		services[index] = service
		p.cache[res.Service.Name] = services
	}
}

// TODO: publish event
func (p *platform) Register(s *registry.Service) error {
	p.Lock()
	defer p.Unlock()

	if err := p.opts.Registry.Register(s); err != nil {
		return err
	}

	service := toProto(s)

	hb := &proto.Heartbeat{
		Id:       s.Nodes[0].Id,
		Service:  service,
		Interval: int64(p.opts.Interval.Seconds()),
		Ttl:      int64((p.opts.Interval.Seconds()) * 5),
	}

	p.heartbeats[hb.Id] = hb

	// now register
	return client.Publish(context.TODO(), client.NewPublication(WatchTopic, &proto.Result{
		Action:    "update",
		Service:   service,
		Timestamp: time.Now().Unix(),
	}))
}

// TODO: publish event
func (p *platform) Deregister(s *registry.Service) error {
	p.Lock()
	defer p.Unlock()

	if err := p.opts.Registry.Deregister(s); err != nil {
		return err
	}

	delete(p.heartbeats, s.Nodes[0].Id)

	// now deregister
	return client.Publish(context.TODO(), client.NewPublication(WatchTopic, &proto.Result{
		Action:    "delete",
		Service:   toProto(s),
		Timestamp: time.Now().Unix(),
	}))
}

func (p *platform) GetService(name string) ([]*registry.Service, error) {
	p.RLock()
	if services, ok := p.cache[name]; ok {
		p.RUnlock()
		return services, nil
	}
	p.RUnlock()

	// disabled discovery?
	if !p.opts.Discovery {
		return p.opts.Registry.GetService(name)
	}

	rsp, err := p.reg.GetService(context.TODO(), &proto2.GetServiceRequest{Service: name})
	if err != nil {
		return nil, err
	}

	var services []*registry.Service
	for _, service := range rsp.Services {
		services = append(services, toService(service))
	}
	return services, nil
}

// TODO: prepoulate the cache
func (p *platform) ListServices() ([]*registry.Service, error) {
	p.RLock()
	if cache := p.cache; len(cache) > 0 {
		p.RUnlock()
		var services []*registry.Service
		for _, service := range cache {
			services = append(services, service...)
		}
		return services, nil
	}
	p.RUnlock()

	// disabled discovery?
	if !p.opts.Discovery {
		return p.opts.Registry.ListServices()
	}

	rsp, err := p.reg.ListServices(context.TODO(), &proto2.ListServicesRequest{})
	if err != nil {
		return nil, err
	}

	var services []*registry.Service
	for _, service := range rsp.Services {
		services = append(services, toService(service))
	}
	return services, nil
}

// TODO: subscribe to events rather than the registry itself?
func (p *platform) Watch() (registry.Watcher, error) {
	// disabled discovery?
	if !p.opts.Discovery {
		return p.opts.Registry.Watch()
	}

	wc, err := p.reg.Watch(context.TODO(), &proto2.WatchRequest{})
	if err != nil {
		return nil, err
	}
	return &watcher{wc}, nil
}

func (p *platform) String() string {
	return "platform"
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.watcher == nil {
		w, err := p.Watch()
		if err != nil {
			return err
		}
		p.watcher = w
		p.exit = make(chan bool)
		go p.run()
	}

	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()
	if p.watcher != nil {
		p.watcher.Stop()
		p.watcher = nil
		close(p.exit)
	}
	return nil
}
