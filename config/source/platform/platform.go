package platform

/*
	Platform source uses the Micro config-srv
*/

import (
	"github.com/micro/go-platform/config"
)

func NewSource(opts ...config.SourceOption) config.Source {
	return config.NewSource(opts...)
}
