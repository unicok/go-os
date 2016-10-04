package platform

import (
	"github.com/micro/go-os/monitor"
)

func NewMonitor(opts ...monitor.Option) monitor.Monitor {
	return monitor.NewMonitor(opts...)
}
