package platform

import (
	"github.com/micro/go-platform/log"
)

func NewLog(opts ...log.Option) log.Log {
	return log.NewLog(opts...)
}
