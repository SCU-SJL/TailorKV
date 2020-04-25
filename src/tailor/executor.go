package tailor

import (
	"sync"
	"time"
)

const (
	setex byte = iota
	setnx
	set
	get
	del
	unlink
	incr
	incrby
	ttl
)

type job struct {
	op   byte
	key  string
	val  interface{}
	exp  time.Duration
	done chan struct{}
	res  response
}

type response struct {
	value interface{}
	ok    bool
	err   error
}

type executor struct {
	c     *Cache
	jobs  chan *job
	size  uint8
	count uint8
	mu    sync.Mutex
}

func newExecutor(c *Cache, size uint8) *executor {
	return &executor{
		c:     c,
		jobs:  make(chan *job, size),
		size:  size,
		count: 0,
	}
}

func (exc *executor) execute(j *job) {
	exc.jobs <- j
}

/*
 * write - serial
 * read  - parallel
 */
func (exc *executor) server() {
	for j := range exc.jobs {
		switch j.op {
		case setex:
			exc.c.setex(j.key, j.val, j.exp)
		case setnx:
			exc.c.setnx(j.key, j.val)
		case set:
			exc.c.set(j.key, j.val)
		case get:
			go func() {
				if exc.isReady() {
					exc.addCount(true)
					j.res.value = exc.c.get(j.key)
					close(j.done)
					exc.addCount(false)
				} else {
					exc.jobs <- j
				}
			}()
		case del:
			exc.c.del(j.key)
		case unlink:
			exc.c.unlink(j.key)
		case incr:
			j.res.err = exc.c.incr(j.key)
		case incrby:
			j.res.err = exc.c.incrby(j.key, j.val.(string))
		case ttl:
			go func() {
				if exc.isReady() {
					exc.addCount(true)
					j.res.value, j.res.ok = exc.c.ttl(j.key)
					close(j.done)
					exc.addCount(false)
				} else {
					exc.jobs <- j
				}
			}()
		}
	}
}

func (exc *executor) addCount(add bool) {
	exc.mu.Lock()
	if add {
		exc.count++
	} else {
		exc.count--
	}
	exc.mu.Unlock()
}

func (exc *executor) isReady() bool {
	exc.mu.Lock()
	defer exc.mu.Unlock()
	return exc.count < exc.size
}
