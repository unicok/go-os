package etcd

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/coreos/etcd/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-platform/sync"

	"golang.org/x/net/context"
)

type etcdLeader struct {
	id   string
	path string
	node *registry.Node
	api  client.KeysAPI
}

type etcdElected struct {
	exit chan bool
}

func (e *etcdLeader) Id() string {
	return e.id
}

func (e *etcdLeader) Leader() (*registry.Node, error) {
	path := path.Join(e.path, strings.Replace(e.id, "/", "-", -1))

	rsp, err := e.api.Get(context.Background(), path, nil)
	if err != nil {
		return nil, err
	}

	var node *registry.Node

	if err := json.Unmarshal([]byte(rsp.Node.Value), &node); err != nil {
		return nil, err
	}

	return node, nil
}

func (e *etcdLeader) Elect() (sync.Elected, error) {
	return &etcdElected{}, nil
}

func (e *etcdLeader) Status() (sync.LeaderStatus, error) {
	return sync.FollowerStatus, nil
}

func (e *etcdElected) Revoked() (chan struct{}, error) {
	ch := make(chan struct{})
	return ch, nil
}

func (e *etcdElected) Resign() error {
	return nil
}
