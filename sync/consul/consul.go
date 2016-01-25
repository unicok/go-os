package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	mtx "sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/micro/go-micro/registry"
	sync "github.com/micro/go-platform/sync"
)

type consulSync struct {
	opts sync.Options
	c    *api.Client
	node *registry.Node
}

type consulLock struct {
	opts sync.LockOptions
	l    *api.Lock
	id   string
}

type consulLeader struct {
	opts sync.LeaderOptions
	c    *api.Client

	id  string
	key string
	// marshalled registry node
	srv []byte

	// used to sync back status
	statusCh chan sync.LeaderStatus

	mtx.Mutex
	status sync.LeaderStatus
}

type consulElected struct {
	ch chan sync.LeaderStatus
	rv <-chan struct{}
	l  *api.Lock
}

func (c *consulLeader) Leader() (*registry.Node, error) {
	kv, _, err := c.c.KV().Get(c.key, nil)
	if err != nil {
		return nil, err
	}
	var node *registry.Node
	if err := json.Unmarshal(kv.Value, &node); err != nil {
		return nil, err
	}
	return node, nil
}

func (c *consulLeader) Id() string {
	return c.id
}

func (c *consulLeader) Elect() (sync.Elected, error) {
	lc, err := c.c.LockOpts(&api.LockOptions{
		Key:   c.key,
		Value: c.srv,
	})
	if err != nil {
		return nil, err
	}

	rv, err := lc.Lock(nil)
	if err != nil {
		return nil, err
	}

	c.statusCh <- sync.ElectedStatus

	// lock acquired
	return &consulElected{
		rv: rv,
		l:  lc,
		ch: c.statusCh,
	}, nil
}

func (c *consulLeader) Status() (sync.LeaderStatus, error) {
	c.Lock()
	defer c.Unlock()
	return c.status, nil
}

func (c *consulElected) Revoked() (chan struct{}, error) {
	select {
	case <-c.rv:
		return nil, errors.New("already revoked")
	default:
	}

	ch := make(chan struct{}, 1)

	go func() {
		st := <-ch
		c.ch <- sync.FollowerStatus
		ch <- st
	}()

	return ch, nil
}

func (c *consulElected) Resign() error {
	c.ch <- sync.FollowerStatus
	err := c.l.Unlock()
	return err
}

func (c *consulLock) Id() string {
	return c.id
}

func (c *consulLock) Acquire() error {
	lc, err := c.l.Lock(nil)
	if err != nil {
		return err
	}

	select {
	case <-lc:
		return errors.New("lock lost")
	default:
	}

	return nil
}

func (c *consulLock) Release() error {
	return c.l.Unlock()
}

func (c *consulSync) Lock(id string, opts ...sync.LockOption) (sync.Lock, error) {
	var options sync.LockOptions
	for _, o := range opts {
		o(&options)
	}

	if options.Wait <= time.Duration(0) {
		options.Wait = api.DefaultLockWaitTime
	}

	ttl := fmt.Sprintf("%v", options.TTL)
	if options.TTL <= time.Duration(0) {
		ttl = api.DefaultLockSessionTTL
	}

	key := filepath.Join(c.opts.Namespace, "lock", id)
	key = strings.Replace(key, "/", "#", -1)

	l, err := c.c.LockOpts(&api.LockOptions{
		Key:          key,
		LockWaitTime: options.Wait,
		SessionTTL:   ttl,
	})

	if err != nil {
		return nil, err
	}

	return &consulLock{
		id:   id,
		opts: options,
		l:    l,
	}, nil
}

func (c *consulSync) Leader(id string, opts ...sync.LeaderOption) (sync.Leader, error) {
	var options sync.LeaderOptions
	for _, o := range opts {
		o(&options)
	}

	key := filepath.Join(c.opts.Namespace, "leader", id)
	key = strings.Replace(key, "/", "#", -1)

	b, err := json.Marshal(c.node)
	if err != nil {
		return nil, err
	}

	cl := &consulLeader{
		c:      c.c,
		id:     id,
		key:    key,
		opts:   options,
		status: sync.FollowerStatus,
		srv:    b,
	}

	ch := make(chan sync.LeaderStatus, 1)

	go func() {
		for status := range ch {
			cl.Lock()
			cl.status = status
			cl.Unlock()
		}
	}()

	cl.statusCh = ch

	return cl, nil
}

func (c *consulSync) Start() error {
	return nil
}

func (c *consulSync) Stop() error {
	return nil
}

func NewSync(opts ...sync.Option) sync.Sync {
	var options sync.Options
	for _, o := range opts {
		o(&options)
	}

	if len(options.Namespace) == 0 {
		options.Namespace = sync.DefaultNamespace
	}

	var node *registry.Node
	if options.Service != nil && len(options.Service.Nodes) > 0 {
		node = options.Service.Nodes[0]
	}

	config := api.DefaultConfig()
	// set host
	// config.Host something
	// check if there are any addrs
	if len(options.Nodes) > 0 {
		addr, port, err := net.SplitHostPort(options.Nodes[0])
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "8500"
			config.Address = fmt.Sprintf("%s:%s", addr, port)
		} else if err == nil {
			config.Address = fmt.Sprintf("%s:%s", addr, port)
		}
	}

	client, _ := api.NewClient(config)

	return &consulSync{
		opts: options,
		c:    client,
		node: node,
	}
}
