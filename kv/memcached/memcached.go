package memcached

import (
	mc "github.com/bradfitz/gomemcache/memcache"
	"github.com/myodc/go-platform/kv"
)

type mkv struct {
	Client *mc.Client
}

func (m *mkv) Get(key string) (*kv.Item, error) {
	keyval, err := m.Client.Get(key)
	if err != nil && err == mc.ErrCacheMiss {
		return nil, kv.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if keyval == nil {
		return nil, kv.ErrNotFound
	}

	return &kv.Item{
		Key:   keyval.Key,
		Value: keyval.Value,
	}, nil
}

func (m *mkv) Del(key string) error {
	return m.Client.Delete(key)
}

func (m *mkv) Put(item *kv.Item) error {
	return m.Client.Set(&mc.Item{
		Key:   item.Key,
		Value: item.Value,
	})
}

func NewKV(addrs []string) kv.KV {
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:11211"}
	}
	return &mkv{
		Client: mc.New(addrs...),
	}
}
