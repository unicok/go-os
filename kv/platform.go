package kv

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
	proto "github.com/micro/go-platform/kv/proto"

	"golang.org/x/net/context"
	"stathat.com/c/consistent"
)

// Consistently hashed in memory key-value store utilising
// all the services in the network. Aww yea.
// Can optionally be namespaced using that provided
type platform struct {
	opts Options

	hash *consistent.Consistent

	sync.RWMutex
	nodes map[string]int64

	running bool
	exit    chan bool
}

type Announcement struct {
	// Which namespace does it belong to
	Namespace string
	Address   string
	Timestamp int64
}

var (
	// not really needed but you know...
	serviceName = "go.micro.kv"

	GossipTopic = "go.micro.kv.announce"
	GossipEvent = time.Second * 5
	ReaperEvent = time.Second * 30
)

func newPlatform(opts ...Option) KV {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Client == nil {
		options.Client = client.DefaultClient
	}

	if options.Server == nil {
		options.Server = server.DefaultServer
	}

	if options.Replicas == 0 {
		options.Replicas = 1
	}

	p := &platform{
		opts:  options,
		hash:  consistent.New(),
		nodes: make(map[string]int64),
	}

	options.Server.Subscribe(
		options.Server.NewSubscriber(
			GossipTopic, p.subscriber, server.InternalSubscriber(true),
		),
	)

	options.Server.Handle(
		options.Server.NewHandler(
			&proto.KV{new(kv)}, server.InternalHandler(true),
		),
	)

	return p
}

func (a *Announcement) Topic() string {
	return GossipTopic
}

func (a *Announcement) Message() interface{} {
	return a
}

func (a *Announcement) ContentType() string {
	return "application/json"
}

func (p *platform) address() string {
	config := p.opts.Server.Options()

	var advt, host string
	var port int

	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	parts := strings.Split(advt, ":")
	if len(parts) > 1 {
		host = strings.Join(parts[:len(parts)-1], ":")
		port, _ = strconv.Atoi(parts[len(parts)-1])
	} else {
		host = parts[0]
	}

	addr, _ := extractAddress(host)

	if port > 0 {
		return fmt.Sprintf("%s:%d", addr, port)
	}

	return addr
}

func (p *platform) reap() {
	t := time.Now().Unix()
	r := int64(GossipEvent.Seconds() * 1.5)

	// reap nodes
	p.Lock()
	for node, seen := range p.nodes {
		// Is last greater than GossipEvent time plus some
		if last := t - seen; last > r {
			delete(p.nodes, node)
			p.hash.Remove(node)
		}
	}
	p.Unlock()

	// reap keys
	mtx.Lock()

	for key, item := range items {
		// don't expire zero or less
		if item.Expiration <= 0 {
			continue
		}

		// Delta greater than expiration
		if delta := t - item.Timestamp; delta > item.Expiration {
			delete(items, key)
		}
	}
	mtx.Unlock()
}

func (p *platform) run(exit chan bool) {
	t := time.NewTicker(GossipEvent)
	r := time.NewTicker(ReaperEvent)

	for {
		select {
		case <-t.C:
			p.publish()
		case <-r.C:
			p.reap()
		case <-exit:
			t.Stop()
			r.Stop()
			return
		}
	}
}

func (p *platform) publish() error {
	a := &Announcement{
		Namespace: p.opts.Namespace,
		Address:   p.address(),
		Timestamp: time.Now().Unix(),
	}
	return p.opts.Client.Publish(context.TODO(), a)
}

func (p *platform) subscriber(ctx context.Context, a *Announcement) error {
	p.Lock()
	defer p.Unlock()

	if p.opts.Namespace != a.Namespace {
		return nil
	}

	_, ok := p.nodes[a.Address]
	if !ok {
		p.hash.Add(a.Address)
	}

	p.nodes[a.Address] = a.Timestamp
	return nil
}

func (p *platform) Get(key string) (*Item, error) {
	nodes, err := p.hash.GetN(key, p.opts.Replicas)
	if err != nil {
		return nil, err
	}

	req := p.opts.Client.NewRequest(serviceName, "KV.Get", &proto.GetRequest{
		Key: key,
	})

	rsp := &proto.GetResponse{}

	for _, node := range nodes {
		// query node and return
		if err := p.opts.Client.CallRemote(context.TODO(), node, req, rsp); err == nil {
			if rsp.Item == nil {
				continue
			}
			return &Item{
				Key:        rsp.Item.Key,
				Value:      rsp.Item.Value,
				Expiration: time.Second * time.Duration(rsp.Item.Expiration),
			}, nil
		}
	}

	return nil, ErrNotFound
}

func (p *platform) Del(key string) error {
	nodes, err := p.hash.GetN(key, p.opts.Replicas)
	if err != nil {
		return err
	}

	req := p.opts.Client.NewRequest(serviceName, "KV.Del", &proto.DelRequest{
		Key: key,
	})

	var gerr error

	for _, node := range nodes {
		rsp := &proto.DelResponse{}
		if err := p.opts.Client.CallRemote(context.TODO(), node, req, rsp); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (p *platform) Put(item *Item) error {
	nodes, err := p.hash.GetN(item.Key, p.opts.Replicas)
	if err != nil {
		return err
	}

	req := p.opts.Client.NewRequest(serviceName, "KV.Put", &proto.PutRequest{
		Item: &proto.Item{
			Key:        item.Key,
			Value:      item.Value,
			Expiration: int64(item.Expiration.Seconds()),
		},
	})

	var gerr error

	for _, node := range nodes {
		rsp := &proto.PutResponse{}
		if err := p.opts.Client.CallRemote(context.TODO(), node, req, rsp); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()

	if p.running {
		return nil
	}

	exit := make(chan bool)
	go p.run(exit)

	p.exit = exit
	p.running = true
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

	return nil
}

func (p *platform) String() string {
	return "platform"
}
