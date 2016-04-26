# KV - key value interface
 
Provides a high level abstraction for key-value stores.

## Interface

```go
type KV interface {
	Close() error
	Get(key string) (*Item, error)
	Del(key string) error
	Put(item *Item) error
	String() string
}

type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
}

func NewKV(opts ...Option) KV {
	return newPlatform(opts...)
}
```

## Supported Backends

- Gossip
- Memcached
- Redis
