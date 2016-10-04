package platform

import (
	"github.com/micro/go-os/kv"
)

func NewKV(opts ...kv.Option) kv.KV {
	return kv.NewKV(opts...)
}
