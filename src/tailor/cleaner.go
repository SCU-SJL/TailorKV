package tailor

import (
	"fmt"
	"time"
)

type cleaner struct {
	interval time.Duration
	stop     chan bool
	clean    func(*cache)
}

func (cl *cleaner) run(c *cache) {
	ticker := time.Tick(cl.interval)
	for {
		select {
		case <-ticker:
			cl.clean(c)
		case <-cl.stop:
			return
		}
	}
}

func (cl *cleaner) stopNow() {
	cl.stop <- true
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
		stop:     make(chan bool),
		clean: func(c *cache) {
			c.delExpired()
		},
	}
	return cl
}
