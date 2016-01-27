package monitor

import (
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/server"
)

type Monitor interface {
	NewHealthChecker(id, desc string, fn HealthCheck) HealthChecker
	Register(HealthChecker) error
	Deregister(HealthChecker) error
	HealthChecks() ([]HealthChecker, error)
	Start() error
	Stop() error
	String() string
}

type HealthChecker interface {
	// Unique id of the healthcheck
	Id() string
	// Description of what the healthcheck does
	Description() string
	// Runs the the healthcheck
	Run() (map[string]string, error)
	// Returns the status of the last run
	// Better where the healthcheck is expensive to run
	Status() (map[string]string, error)
}

type HealthCheck func() (map[string]string, error)

type Option func(*Options)

type Options struct {
	Interval time.Duration
	Client   client.Client
	Server   server.Server
}

var (
	HealthCheckTopic = "micro.monitor.healthcheck"
)

func NewMonitor(opts ...Option) Monitor {
	return newPlatform(opts...)
}
