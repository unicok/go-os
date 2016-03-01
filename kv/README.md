# KV - key value interface
 
Provides a high level abstraction for key-value stores.

## Interface

```go
type KV interface {
	Get(key string) (*Item, error)
	Del(key string) error
	Put(item *Item) error

	// Runtime. Could be used for internal reaping
	// of expired keys or publishing info, gossip,
	// etc
	Start() error
	Stop() error
	// Name
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
