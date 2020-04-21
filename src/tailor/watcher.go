package tailor

import (
	"time"
)

type watcher struct {
	interval time.Duration
	stop     chan struct{}
	op       func(c *Cache)
	stopped  chan struct{}
}

func (w *watcher) run(c *Cache) {
	defer func() {
		close(w.stopped)
	}()
	w.stop = make(chan struct{})
	w.stopped = make(chan struct{})
	ticker := time.Tick(w.interval)
	for {
		select {
		case <-ticker:
			if w.op != nil {
				// cannot and no need create a goroutine here,
				// as the watcher may be stopped or replaced at any time.
				w.op(c)
			}
		case <-w.stop:
			return
		}
	}
}

func newWatcher(t time.Duration, f func(*Cache)) *watcher {
	return &watcher{
		interval: t,
		stop:     make(chan struct{}),
		op:       f,
	}
}
