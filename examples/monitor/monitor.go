package main

import (
	"errors"
	"fmt"
	"time"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-platform/monitor"
)

// Return successful healthcheck
func success() (map[string]string, error) {
	return map[string]string{
		"msg": "a successful check",
		"foo": "bar",
		"metric": "1",
		"label": "cruft",
		"stats": "123.0",
	}, nil
}

// Return failing healthcheck
func failure() (map[string]string, error) {
	return map[string]string{
		"msg": "a catastrophic failure occurred",
		"foo": "ugh",
		"metric": "-0.0001",
		"label": "",
		"stats": "NaN",
	}, errors.New("Unknown exception")
}

func main() {
	cmd.Init()
	m := monitor.NewMonitor(
		monitor.Interval(time.Second),
	)

	hc1 := m.NewHealthChecker("go.micro.healthcheck.ping", "This is a ping healthcheck that succeeds", success)
	hc2 := m.NewHealthChecker("go.micro.healthcheck.pong", "This is a pong healthcheck that fails", failure)

	m.Register(hc1)
	m.Register(hc2)

	fmt.Println("Starting monitor, will sleep for 10 seconds")
	m.Start()
	<-time.After(time.Second * 10)
	fmt.Println("Stopping monitor")
	m.Stop()
}
