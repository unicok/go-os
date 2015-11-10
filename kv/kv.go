package kv

type KV interface {
	Get(key string) (*Item, error)
	Del(key string) error
	Put(item *Item) error
}

type Item struct {
	Key   string
	Value []byte
}
