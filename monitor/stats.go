package monitor

import (
	"syscall"
)

type stats struct {
	utime   syscall.Timeval
	stime   syscall.Timeval
	maxRss  int64
	inBlock int64
	ouBlock int64
	volCtx  int64
	invCtx  int64
}

func newStats() *stats {
	var r syscall.Rusage

	if err := syscall.Getrusage(0, &r); err != nil {
		return nil
	}

	return &stats{
		utime:   r.Utime,
		stime:   r.Stime,
		maxRss:  r.Maxrss,
		inBlock: r.Inblock,
		ouBlock: r.Oublock,
		volCtx:  r.Nvcsw,
		invCtx:  r.Nivcsw,
	}
}
