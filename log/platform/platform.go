package platform

import (
	"github.com/micro/go-os/log"
)

func NewLog(opts ...log.Option) log.Log {
	return log.NewLog(opts...)
}
