package kv

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

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

type Option func(o *Options)

func NewKV(opts ...Option) KV {
	return newPlatform(opts...)
}
