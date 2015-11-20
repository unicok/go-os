package redis

import (
	"github.com/micro/go-platform/kv"
	redis "gopkg.in/redis.v3"
)

type rkv struct {
	Client *redis.Client
}

func (r *rkv) Get(key string) (*kv.Item, error) {
	val, err := r.Client.Get(key).Bytes()

	if err != nil && err == redis.Nil {
		return nil, kv.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, kv.ErrNotFound
	}

	return &kv.Item{
		Key:   key,
		Value: val,
	}, nil
}

func (r *rkv) Del(key string) error {
	return r.Client.Del(key).Err()
}

func (r *rkv) Put(item *kv.Item) error {
	return r.Client.Set(item.Key, item.Value, 0).Err()
}

func NewKV(addr string) kv.KV {
	if len(addr) == 0 {
		addr = "127.0.0.1:6379"
	}

	return &rkv{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
	}
}
