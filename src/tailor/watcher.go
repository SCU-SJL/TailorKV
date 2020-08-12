package tailor

import (
	"sync"
	"time"
)

type watcher struct {
	interval time.Duration
	stop     chan struct{}
	op       func(c *Cache)
	waitStop chan struct{}
	mu       sync.Mutex
}

func (w *watcher) run(c *Cache) {
	w.stop = make(chan struct{})
	w.waitStop = make(chan struct{})
	ticker := time.Tick(w.interval)

	defer func() {
		close(w.waitStop)
	}()

	for {
		select {
		case <-ticker:
			if w.op != nil {
				// cannot and no need create a goroutine here,
				// as the watcher may be waitStop or replaced at any time.
				w.op(c)
			}
		case <-w.stop:
			return
		}
	}
}

func (w *watcher) stopNow() {
	close(w.stop)
}

func (w *watcher) stopAndWait() {
	close(w.stop)
	<-w.waitStop
}

func newWatcher(t time.Duration, f func(*Cache)) *watcher {
	return &watcher{
		interval: t,
		stop:     make(chan struct{}),
		op:       f,
		waitStop: make(chan struct{}),
	}
}
