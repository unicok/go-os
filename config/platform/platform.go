package platform

import (
	"github.com/micro/go-os/config"
)

func NewConfig(opts ...config.Option) config.Config {
	return config.NewConfig(opts...)
}
