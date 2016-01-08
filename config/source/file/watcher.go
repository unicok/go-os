package file

import (
	"sync"

	"github.com/micro/go-platform/config"
	"gopkg.in/fsnotify.v1"
)

type watcher struct {
	f *file

	sync.Mutex
	watchers []*watch
}

type watch struct {
	fw   *fsnotify.Watcher
	ch   chan *config.ChangeSet
	exit chan bool
}

func (w *watcher) Changes() <-chan *config.ChangeSet {
	w.Lock()
	defer w.Unlock()

	// do something about this. Maybe create the watcher
	// before hand
	fw, _ := fsnotify.NewWatcher()

	aw := &watch{
		fw:   fw,
		ch:   make(chan *config.ChangeSet),
		exit: make(chan bool),
	}

	aw.fw.Add(w.f.opts.Name)

	go func() {
		// cleanup func
		defer func() {
			aw.fw.Close()
			close(aw.ch)
		}()

		for {
			select {
			case <-aw.fw.Events:
				c, err := w.f.Read()
				if err != nil {
					return
				}
				// send changeset
				aw.ch <- c
			case <-aw.fw.Errors:
				// exit on err
				return
			case <-aw.exit:
				// told to exit
				return
			}
		}
	}()

	w.watchers = append(w.watchers, aw)

	return aw.ch
}

func (w *watcher) Stop() error {
	w.Lock()
	defer w.Unlock()

	for _, w := range w.watchers {
		close(w.exit)
	}

	w.watchers = nil

	return nil
}
