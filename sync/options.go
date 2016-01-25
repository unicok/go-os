package sync

import (
	"time"
)

type LockOptions struct {
	Id   string
	Ttl  time.Duration
	Wait time.Duration
}

type LeaderOptions struct{}
