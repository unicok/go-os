package platform

import (
	"github.com/micro/go-platform/kv"
)

func NewKV(opts ...kv.Option) kv.KV {
	return kv.NewKV(opts...)
}
