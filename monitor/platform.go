package monitor

import (
	"errors"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
	proto "github.com/micro/go-platform/monitor/proto"

	"golang.org/x/net/context"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAvailable  = errors.New("not available")
)

type platform struct {
	opts              Options
	name, version, id string

	sync.Mutex
	running bool
	exit    chan bool
	hc      map[string]HealthChecker
}

func newPlatform(opts ...Option) Monitor {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}

	if opt.Client == nil {
		opt.Client = client.DefaultClient
	}

	if opt.Server == nil {
		opt.Server = server.DefaultServer
	}

	if opt.Interval == time.Duration(0) {
		opt.Interval = time.Minute
	}

	c := opt.Server.Config()

	return &platform{
		name:    c.Name(),
		version: c.Version(),
		id:      c.Id(),
		opts:    opt,
		exit:    make(chan bool, 1),
		hc:      make(map[string]HealthChecker),
	}
}

func (p *platform) update(h HealthChecker) {
	res, err := h.Run()
	status := proto.HealthCheck_OK
	errDesc := ""
	if err != nil {
		status = proto.HealthCheck_ERROR
		errDesc = err.Error()
	}

	hcProto := &proto.HealthCheck{
		Id:          h.Id(),
		Description: h.Description(),
		Timestamp:   time.Now().Unix(),
		Service: &proto.HealthCheck_Service{
			Name:    p.name,
			Version: p.version,
			Id:      p.id,
		},
		Interval: int64(p.opts.Interval.Seconds()),
		Ttl:      3600,
		Status:   status,
		Results:  res,
		Error:    errDesc,
	}

	req := p.opts.Client.NewPublication(HealthCheckTopic, hcProto)
	p.opts.Client.Publish(context.TODO(), req)
}

func (p *platform) run() {
	t := time.NewTicker(p.opts.Interval)

	for {
		select {
		case <-t.C:
			p.Lock()
			for _, check := range p.hc {
				go p.update(check)
			}
			p.Unlock()
		case <-p.exit:
			t.Stop()
			return
		}
	}
}

func (p *platform) NewHealthChecker(id, desc string, hc HealthCheck) HealthChecker {
	return newHealthChecker(id, desc, hc)
}

func (p *platform) Register(hc HealthChecker) error {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.hc[hc.Id()]; ok {
		return ErrAlreadyExists
	}
	p.hc[hc.Id()] = hc
	return nil
}

func (p *platform) Deregister(hc HealthChecker) error {
	p.Lock()
	defer p.Unlock()
	delete(p.hc, hc.Id())
	return nil
}

func (p *platform) HealthChecks() ([]HealthChecker, error) {
	var hcs []HealthChecker
	p.Lock()
	for _, hc := range p.hc {
		hcs = append(hcs, hc)
	}
	p.Unlock()
	return hcs, nil
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()
	if p.running {
		return nil
	}
	go p.run()
	p.running = true
	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()
	if p.running {
		p.exit <- true
		p.running = false
	}
	return nil
}
