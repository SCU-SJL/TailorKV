package tailor

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	neCache  *cache
	exCache  *cache
	watcher  *watcher
	cleaner  *cleaner
	watchMu  sync.Mutex
	wStopped bool
}

func NewCache(defaultExpiration time.Duration, m map[string]Item) *Cache {
	var nec, exc *cache
	if defaultExpiration <= 0 { // if expiry time is not greater than zero, then make it NoExpiration
		nec = newCache(NoExpiration, m)
		exc = newCache(NoExpiration, m)
	} else {
		exc = newCache(defaultExpiration, m)
		nec = exc
	}
	// clean expired data twice each second
	cl := defaultCleaner(500 * time.Millisecond)
	C := &Cache{
		neCache:  nec,
		exCache:  exc,
		cleaner:  cl,
		wStopped: true,
	}
	go cl.run(exc)
	return C
}

// thread-safe
func (c *Cache) ReplaceDaemonOp(wi time.Duration, f func(*Cache)) {
	if c.watcher != nil {
		_ = c.StopWatchingSync()
		c.watcher.interval = wi
		c.watcher.op = f
	} else {
		c.watchMu.Lock()
		c.watcher = newWatcher(wi, f)
		c.watchMu.Unlock()
	}
}

// This func may cause blocking when the daemon op is very complicated.
// Use StopWatchingAsync() if you don't want to get blocked.
func (c *Cache) StopWatchingSync() error {
	c.watchMu.Lock()
	if c.watcher == nil {
		c.watchMu.Unlock()
		return fmt.Errorf("there is no daemon watcher")
	}
	if c.wStopped {
		c.watchMu.Unlock()
		return fmt.Errorf("daemon watcher has stopped")
	}
	c.watcher.stopAndWait()
	c.wStopped = true
	c.watchMu.Unlock()
	return nil
}

func (c *Cache) StopWatchingAsync() error {
	c.watchMu.Lock()
	if c.watcher == nil {
		c.watchMu.Unlock()
		return fmt.Errorf("there is no daemon watcher")
	}
	if c.wStopped {
		c.watchMu.Unlock()
		return fmt.Errorf("daemon watcher has stopped")
	}
	go func() {
		c.watcher.stopAndWait()
		c.wStopped = true
		c.watchMu.Unlock()
	}()
	return nil
}

func (c *Cache) StartWatching() error {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()
	if c.watcher != nil {
		if !c.wStopped {
			return fmt.Errorf("the daemon watcher has started")
		} else {
			go c.watcher.run(c)
		}
	}
	c.wStopped = false
	return nil
}

func (c *Cache) AddDelHandler(f func(key string, val interface{})) {
	c.neCache.addDelHandler(f)
	c.exCache.addDelHandler(f)
}

func (c *Cache) set(key string, val interface{}) {
	if _, ok := c.exCache.get(key); ok {
		c.exCache.set(key, val, DefaultExpiration)
	} else {
		c.neCache.set(key, val, DefaultExpiration)
	}
}

func (c *Cache) setnx(key string, val interface{}) bool {
	err := c.neCache.setnx(key, val, DefaultExpiration)
	return err == nil
}

func (c *Cache) setex(key string, val interface{}, t time.Duration) {
	c.exCache.set(key, val, t)
}

func (c *Cache) get(key string) interface{} {
	val, ok := c.neCache.get(key)
	if ok {
		return val
	}
	val, ok = c.exCache.get(key)
	if ok {
		return val
	}
	return nil
}

func (c *Cache) del(key string) {
	c.neCache.del(key)
	if c.exCache != c.neCache {
		c.exCache.del(key)
	}
}

func (c *Cache) unlink(key string) {
	c.neCache.unlink(key)
	if c.exCache != c.neCache {
		c.exCache.unlink(key)
	}
}

func (c *Cache) incr(key string) error {
	err := c.neCache.incrby(key, 1)
	if err != nil {
		return c.exCache.incrby(key, 1)
	}
	return err
}

func (c *Cache) incrby(key string, s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	err = c.neCache.incrby(key, n)
	if err != nil {
		return c.exCache.incrby(key, n)
	}
	return err
}

func (c *Cache) Keys(exp string) ([]KV, error) {
	var res []KV
	res, err := c.neCache.keys(exp)
	if err != nil {
		return nil, err
	}
	exKV, err := c.exCache.keys(exp)
	if err != nil {
		return nil, err
	}
	res = append(res, exKV...)
	return res, nil
}

func (c *Cache) ttl(key string) (time.Duration, bool) {
	return c.exCache.ttl(key)
}

// TODO enhancement required in the future
func (c *Cache) Save(filename string, ok chan bool) {
	go func() {
		err := c.neCache.saveFile("NE" + filename)
		if err != nil {
			ok <- false
		}
		ok <- true
	}()

	go func() {
		err := c.exCache.saveFile("EX" + filename)
		if err != nil {
			ok <- false
		}
		ok <- true
	}()
}

func (c *Cache) Load(filename string) error {
	err := c.neCache.loadFile("NE" + filename)
	if err != nil {
		return err
	}
	err = c.exCache.loadFile("EX" + filename)
	return err
}

func (c *Cache) Cls() {
	c.exCache.cls()
	c.neCache.cls()
}

func (c *Cache) Cnt() int {
	// the result contains the expired items
	// which are not cleaned when the func is called.
	return c.exCache.cnt() + c.neCache.cnt()
}
