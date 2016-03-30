package platform

import (
	"github.com/micro/go-platform/config"
)

func NewConfig(opts ...config.Option) config.Config {
	return config.NewConfig(opts...)
}
