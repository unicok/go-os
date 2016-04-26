package monitor

import (
	"errors"
	"runtime"
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
	exit              chan bool
	opts              Options
	name, version, id string

	sync.Mutex
	hc   map[string]HealthChecker
	stat *stats
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

	c := opt.Server.Options()

	p := &platform{
		name:    c.Name,
		version: c.Version,
		id:      c.Id,
		opts:    opt,
		exit:    make(chan bool),
		hc:      make(map[string]HealthChecker),
	}

	go p.run()
	return p
}

func (p *platform) stats() {
	// nothing to diff
	// defer publishing
	if p.stat == nil {
		s := newStats()
		if s == nil {
			return
		}
		p.stat = s
		return
	}

	o := p.stat
	s := newStats()

	// update
	p.stat = s

	cpu := &proto.CPU{
		UserTime:     uint64(s.utime.Nano() - o.utime.Nano()),
		SystemTime:   uint64(s.stime.Nano() - o.stime.Nano()),
		VolCtxSwitch: uint64(s.volCtx - o.volCtx),
		InvCtxSwitch: uint64(s.invCtx - o.invCtx),
	}

	memory := &proto.Memory{
		MaxRss: uint64(s.maxRss),
	}

	disk := &proto.Disk{
		InBlock: uint64(s.inBlock - o.inBlock),
		OuBlock: uint64(s.ouBlock - o.ouBlock),
	}

	rm := runtime.MemStats{}
	runtime.ReadMemStats(&rm)

	rtime := &proto.Runtime{
		NumThreads: uint64(runtime.NumGoroutine()),
		HeapTotal:  rm.HeapAlloc,
		HeapInUse:  rm.HeapInuse,
	}

	statsProto := &proto.Stats{
		Service: &proto.Service{
			Name:    p.name,
			Version: p.version,
			Nodes: []*proto.Node{&proto.Node{
				Id: p.id,
			}},
		},
		Interval:  int64(p.opts.Interval.Seconds()),
		Timestamp: time.Now().Unix(),
		Ttl:       3600,
		Cpu:       cpu,
		Memory:    memory,
		Disk:      disk,
		Runtime:   rtime,
		// TODO: add endpoint stats
	}

	req := p.opts.Client.NewPublication(StatsTopic, statsProto)
	p.opts.Client.Publish(context.TODO(), req)
}

func (p *platform) status(status proto.Status_Status) {
	statusProto := &proto.Status{
		Status: status,
		Service: &proto.Service{
			Name:    p.name,
			Version: p.version,
			Nodes: []*proto.Node{&proto.Node{
				Id: p.id,
			}},
		},
		Interval:  int64(p.opts.Interval.Seconds()),
		Timestamp: time.Now().Unix(),
		Ttl:       3600,
	}

	req := p.opts.Client.NewPublication(StatusTopic, statusProto)
	p.opts.Client.Publish(context.TODO(), req)
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
		Service: &proto.Service{
			Name:    p.name,
			Version: p.version,
			Nodes: []*proto.Node{&proto.Node{
				Id: p.id,
			}},
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
	// publish started status
	p.status(proto.Status_STARTED)

	t := time.NewTicker(p.opts.Interval)

	for {
		select {
		case <-t.C:
			// publish stats
			p.stats()
			// publish status
			p.status(proto.Status_RUNNING)
			// publish healthchecks
			p.Lock()
			for _, check := range p.hc {
				go p.update(check)
			}
			p.Unlock()
		case <-p.exit:
			// stop the ticker
			t.Stop()
			// publish started status
			p.status(proto.Status_STOPPED)
			return
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

func (p *platform) RecordStat(r Request, d time.Duration, err error) {
	// TODO: implement recording
	return
}

func (p *platform) String() string {
	return "platform"
}
