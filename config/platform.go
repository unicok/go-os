package config

import (
	"sync"
	"time"

	log "github.com/golang/glog"
)

type platform struct {
	opts Options

	sync.RWMutex
	cset *ChangeSet
	vals Values

	running bool
	exit    chan bool
}

func newPlatform(opts ...Option) Config {
	options := Options{
		PollInterval: DefaultPollInterval,
		Reader:       NewReader(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &platform{
		opts: options,
	}
}

func (p *platform) run(ch chan bool) {
	t := time.NewTicker(p.opts.PollInterval)

	for {
		select {
		case <-t.C:
			p.sync()
		case <-ch:
			t.Stop()
			return
		}
	}
}

func (p *platform) loaded() bool {
	var loaded bool
	p.RLock()
	if p.vals != nil {
		loaded = true
	}
	p.RUnlock()
	return loaded
}

// sync loads all the sources, calls the parser and updates the config
func (p *platform) sync() {
	if len(p.opts.Sources) == 0 {
		log.Errorf("Zero sources available to sync")
		return
	}

	var sets []*ChangeSet

	for _, source := range p.opts.Sources {
		ch, err := source.Read()
		// should we actually skip failing sources?
		// best effort merging right? but what if we
		// already have good config? that would be screwed
		if err != nil {
			p.RLock()
			vals := p.vals
			p.RUnlock()

			// if we have no config, we're going to try
			// load something
			if vals == nil {
				log.Errorf("Failed to load a source %v but current config is empty so continuing", err)
				continue
			} else {
				log.Errorf("Failed to load a source %v backing off", err)
				return
			}
		}
		sets = append(sets, ch)
	}

	set, err := p.opts.Reader.Parse(sets...)
	if err != nil {
		log.Errorf("Failed to parse ChangeSets %v", err)
		return
	}

	p.Lock()
	p.vals, _ = p.opts.Reader.Values(set)
	p.cset = set
	p.Unlock()
}

func (p *platform) Get(path ...string) Value {
	if !p.loaded() {
		p.sync()
	}

	p.Lock()
	defer p.Unlock()

	// did sync actually work?
	if p.vals != nil {
		return p.vals.Get(path...)
	}

	ch := p.cset

	// we are truly screwed, trying to load in a hacked way
	v, err := p.opts.Reader.Values(ch)
	if err != nil {
		log.Errorf("Failed to read values %v trying again", err)
		// man we're so screwed
		// Let's try hack this
		// We should really be better
		if ch == nil || ch.Data == nil {
			ch = &ChangeSet{
				Timestamp: time.Now(),
				Source:    p.String(),
				Data:      []byte(`{}`),
			}
		}
		v, _ = p.opts.Reader.Values(ch)
	}

	// lets set it just because
	p.vals = v

	if p.vals != nil {
		return p.vals.Get(path...)
	}

	// ok we're going hardcore now
	return newValue(nil)
}

func (p *platform) Options() Options {
	return p.opts
}

func (p *platform) Start() error {
	p.Lock()
	defer p.Unlock()
	if p.running {
		return nil
	}

	p.running = true
	p.exit = make(chan bool)
	p.run(p.exit)
	return nil
}

func (p *platform) Stop() error {
	p.Lock()
	defer p.Unlock()
	if !p.running {
		return nil
	}

	p.running = false
	close(p.exit)
	p.exit = nil
	return nil
}

func (p *platform) String() string {
	return "platform"
}
