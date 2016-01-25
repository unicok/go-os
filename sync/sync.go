package sync

import (
	"github.com/micro/go-micro/registry"
)

const (
	FollowerStatus  LeaderStatus = 0
	CandidateStatus LeaderStatus = 1
	ElectedStatus   LeaderStatus = 2
)

type LeaderStatus int32

type Sync interface {
	// distributed lock interface
	Lock(...LockOption) (Lock, error)
	// leader election interface
	Leader(...LeaderOption) (Leader, error)
	// Start/Stop the internal publisher
	// used to announce this client and
	// subscribe to announcements.
	Start() error
	Stop() error
}

type Lock interface {
	Id() string
	Acquire() error
	Release() error
}

type Leader interface {
	// Returns the current leader
	Leader() (*registry.Node, error)
	// Elect self to become leader
	Elect() (Elected, error)
	// Returns the status of this node
	Status() (LeaderStatus, error)
}

type Elected interface {
	// Returns a channel which indicates
	// when the leadership is revoked
	Revoked() (chan bool, error)
	// Resign the leadership
	Resign() error
}

type LockOption func(o *LockOptions)

type LeaderOption func(o *LeaderOptions)
