# KV - key value interface
 
Provides a high level abstraction for key-value stores.

## Interface

```go
type KV interface {
	Get(key string) (*Item, error)
	Del(key string) error
	Put(item *Item) error
}

type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
}
```

## Supported KV Stores

- Platform
- Memcached
- Redis
