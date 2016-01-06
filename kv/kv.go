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
}

type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
}

type Option func(o *Options)

