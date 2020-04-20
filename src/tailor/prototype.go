package tailor

import (
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	neCache *cache
	exCache *cache
	watcher *watcher
	cleaner *cleaner
	watchMu sync.Mutex
}

func NewCache(defaultExpiration time.Duration, m map[string]Item) *Cache {
	var exc *cache
	nec := newCache(defaultExpiration, m)
	if defaultExpiration < 0 {
		exc = newCache(defaultExpiration, m)
	} else {
		exc = nec
	}
	// clean expired data twice each second
	cl := defaultCleaner(500 * time.Millisecond)
	C := &Cache{
		neCache: nec,
		exCache: exc,
		cleaner: cl,
		watcher: nil,
	}
	go cl.run(exc)
	return C
}

func (c *Cache) ReplaceDaemonOp(wi time.Duration, f func(*Cache)) {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()
	if c.watcher != nil {
		c.StopWatching()
		c.watcher.interval = wi
		c.watcher.op = f
	} else {
		c.watcher = newWatcher(wi, f)
	}
	c.StartWatching()
}

func (c *Cache) StopWatching() {
	c.watcher.stop <- true
	time.Sleep(100 * time.Millisecond)
}

func (c *Cache) StartWatching() {
	go c.watcher.run(c)
}

func (c *Cache) AddDelHandler(f func(key string, val interface{})) {
	c.neCache.addDelHandler(f)
	c.exCache.addDelHandler(f)
}

func (c *Cache) Set(key string, val interface{}) {
	c.neCache.set(key, val, DefaultExpiration)
}

func (c *Cache) Setnx(key string, val interface{}) bool {
	err := c.neCache.setnx(key, val, DefaultExpiration)
	return err == nil
}

func (c *Cache) Setex(key string, val interface{}, t time.Duration) {
	c.exCache.set(key, val, t)
}

func (c *Cache) Get(key string) interface{} {
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

func (c *Cache) Del(key string) {
	c.neCache.del(key)
}

func (c *Cache) Incr(key string) error {
	return c.neCache.incrby(key, 1)
}

func (c *Cache) Incrby(key string, s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	return c.neCache.incrby(key, n)
}

func (c *Cache) Keys(exp string) ([]KV, error) {
	return c.neCache.keys(exp)
}

func (c *Cache) Ttl(key string) (time.Duration, bool) {
	return c.neCache.ttl(key)
}

func (c *Cache) Save(filename string) error {
	return c.neCache.saveFile(filename)
}

func (c *Cache) Load(filename string) error {
	return c.neCache.loadFile(filename)
}
