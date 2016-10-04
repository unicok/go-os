package platform

import (
	"github.com/micro/go-os/metrics"
)

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return metrics.NewMetrics(opts...)
}
