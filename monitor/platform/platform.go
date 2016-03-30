package platform

import (
	"github.com/micro/go-platform/monitor"
)

func NewMonitor(opts ...monitor.Option) monitor.Monitor {
	return monitor.NewMonitor(opts...)
}
