package prototype

import "time"

type watcher struct {
	interval time.Duration
	stop     chan bool
	op       func(c *Cache)
	stopped  chan bool
}

func (w *watcher) run(c *Cache) {
	defer func() {
		w.stopped <- true
	}()
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
		stop:     make(chan bool),
		op:       f,
	}
}
