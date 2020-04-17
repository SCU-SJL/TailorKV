package prototype

import "time"

type watcher struct {
	interval time.Duration
	stop     chan bool
	op       func(c *Cache)
}

func (w *watcher) run(c *Cache) {
	ticker := time.Tick(w.interval)
	for {
		select {
		case <-ticker:
			if w.op != nil {
				// cannot create a goroutine here,
				// as the watcher may be stopped or replaced at any time.
				w.op(c)
			}
		case <-w.stop:
			return
		}
	}
}
