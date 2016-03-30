package platform

import (
	"github.com/micro/go-platform/metrics"
)

func NewMetrics(opts ...metrics.Option) metrics.Metrics {
	return metrics.NewMetrics(opts...)
}
