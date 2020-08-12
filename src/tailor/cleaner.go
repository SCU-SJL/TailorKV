package tailor

import (
	"fmt"
	"time"
)

type cleaner struct {
	interval time.Duration
	stop     chan struct{}
	stopped  bool
	clean    func(*cache)
}

func (cl *cleaner) isStopped() bool {
	return cl.stopped
}

func (cl *cleaner) run(c *cache) {
	cl.stopped = false
	ticker := time.Tick(cl.interval)
	for {
		select {
		case <-ticker:
			cl.clean(c)
		case <-cl.stop:
			cl.stopped = true
			return
		}
	}
}

func (cl *cleaner) stopNow() {
	close(cl.stop)
}

func (cl *cleaner) setInterval(t time.Duration) error {
	if t < 0 {
		return fmt.Errorf("interval must greater than zero: %v", t)
	}
	cl.interval = t
	return nil
}

func defaultCleaner(t time.Duration) *cleaner {
	cl := &cleaner{
		interval: t,
		stop:     make(chan struct{}),
		stopped:  false,
		clean: func(c *cache) {
			c.delExpired()
		},
	}
	return cl
}

func newCleanerWithHandler(t time.Duration, handler func(c *cache)) *cleaner {
	cl := &cleaner{
		interval: t,
		stop:     make(chan struct{}),
		stopped:  true,
		clean:    handler,
	}
	return cl
}
