package main

import (
	"fmt"
	"time"

	"github.com/micro/go-platform/sync"
	"github.com/micro/go-platform/sync/consul"
)

func leaderStatus(i int, s sync.Leader, msg string) {
	status, err := s.Status()
	if err != nil {
		fmt.Printf("[leader:%d] error getting leader status %v", err)
		return
	}
	fmt.Printf("[leader:%d] [status:%v] %s\n", i, status, msg)
}

func acquire(i int, s sync.Sync) {
	l, err := s.Lock("alock")
	if err != nil {
		fmt.Printf("[lock:%d] err acquiring lock interface %v\n", i, err)
		return
	}
	fmt.Printf("[lock:%d] attempting to acquire lock\n", i)
	if err := l.Acquire(); err != nil {
		fmt.Printf("[lock:%d] err acquiring lock %v\n", i, err)
		return
	}
	fmt.Printf("[lock:%d] acquired lock!\n", i)
	time.Sleep(time.Second)
	fmt.Printf("[lock:%d] unlocking now\n", i)
	if err := l.Release(); err != nil {
		fmt.Printf("[lock:%d] err releasing lock %v\n", i, err)
	}
	fmt.Printf("[lock:%d] unlocked\n", i)
}

func leader(i int, s sync.Sync) {
	// Get a leader interface
	l, err := s.Leader("king")
	if err != nil {
		fmt.Printf("[leader:%d] err acquiring leader interface %v\n", i, err)
		return
	}

	leaderStatus(i, l, "attempting to elect self")

	// Attempt to elect self
	throne, err := l.Elect()
	if err != nil {
		fmt.Printf("[leader:%d] err electing self %v\n", i, err)
		return
	}

	leaderStatus(i, l, "elected as leader")

	j := 0

	revoked, err := throne.Revoked()
	if err != nil {
		fmt.Printf("[leader:%d] throne already revoked, gnarly %v\n", i, err)

	}

loop:
	for {
		select {
		// Check if we've been revoked
		case <-revoked:
			leaderStatus(i, l, "leadership revoked")
			return
		default:
			leaderStatus(i, l, "I'm leading son")
			time.Sleep(time.Second)
		}

		if j >= 3 {
			break loop
		}

		j++
	}

	leaderStatus(i, l, "resigning now")

	// Resign leadership status
	if err := throne.Resign(); err != nil {
		leaderStatus(i, l, fmt.Sprintf("err resigning %v", err))
	}

	leaderStatus(i, l, "resigned")
}

func main() {
	for i := 0; i < 2; i++ {
		go acquire(i, consul.NewSync())
	}
	acquire(2, consul.NewSync())
	for i := 0; i < 2; i++ {
		go leader(i, consul.NewSync())
	}
	leader(2, consul.NewSync())
	time.Sleep(time.Second * 5)
}
