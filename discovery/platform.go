package discovery

import (
	"sync"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/registry"
)

type platform struct {
	opts  Options
	exit  chan bool
	watch registry.Watcher

	sync.RWMutex
	cache map[string][]*registry.Service
}

func (p *platform) run() {
	p.Lock()
	watch := p.watch
	p.Unlock()

	ch := make(chan *registry.Result)

	go func() {
		for {
			next, err := watch.Next()
			if err != nil {
				close(ch)
				return
			}
			ch <- next
		}
	}()

	for {
		select {
		case <-p.exit:
			return
		case next := <-ch:
			p.update(next)
		}
	}
}

func (p *platform) update(res *registry.Result) {
	p.Lock()
	defer p.Unlock()

	services, ok := p.cache[res.Service.Name]
	if !ok {
		// no service found, let's not go through a convoluted process
		if res.Action == "create" || res.Action == "update" {
			p.cache[res.Service.Name] = []*registry.Service{res.Service}
		}
		return
	}

	// existing service found
	var service *registry.Service
	for _, s := range services {
		if s.Version == res.Service.Version {
			service = s
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

		service = res.Service
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

		service.Nodes = nodes

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
		}
	}
}

// TODO: publish event
func (p *platform) Register(s *registry.Service) error {
	return p.opts.Registry.Register(s)
}

// TODO: publish event
func (p *platform) Deregister(s *registry.Service) error {
	return p.opts.Registry.Deregister(s)
}

func (p *platform) GetService(name string) ([]*registry.Service, error) {
	p.RLock()
	if services, ok := p.cache[name]; ok {
		p.RUnlock()
		return services, nil
	}
	p.RUnlock()

	return p.opts.Registry.GetService(name)
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

	return p.opts.Registry.ListServices()
}

// TODO: subscribe to events rather than the registry itself?
func (p *platform) Watch() (registry.Watcher, error) {
	return p.opts.Registry.Watch()
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()
	if p.watch == nil {
		w, err := p.opts.Registry.Watch()
		if err != nil {
			return err
		}
		p.watch = w
		p.exit = make(chan bool)
		go p.run()
	}

	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()
	if p.watch != nil {
		p.watch.Stop()
		p.watch = nil
		close(p.exit)
	}
	return nil
}

func newPlatform(opts ...Option) Discovery {
	var opt Options

	for _, o := range opts {
		o(&opt)
	}

	if opt.Registry == nil {
		opt.Registry = registry.DefaultRegistry
	}

	if opt.Broker == nil {
		opt.Broker = broker.DefaultBroker
	}

	return &platform{
		opts:  opt,
		cache: make(map[string][]*registry.Service),
	}
}
