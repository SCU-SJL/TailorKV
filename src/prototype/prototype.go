package prototype

import (
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	cache   *cache
	watcher *watcher
	cleaner *cleaner
	watchMu sync.Mutex
}

func NewCache(defaultExpiration time.Duration, m map[string]Item) *Cache {
	c := newCache(defaultExpiration, m)
	cl := defaultCleaner(1 * time.Second)
	C := &Cache{
		cache:   c,
		cleaner: cl,
		watcher: nil,
	}
	go cl.run(c)
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
	c.cache.addDelHandler(f)
}

func (c *Cache) Set(key string, val interface{}) {
	c.cache.set(key, val, DefaultExpiration)
}

func (c *Cache) Setnx(key string, val interface{}) bool {
	err := c.cache.setnx(key, val, DefaultExpiration)
	return err == nil
}

func (c *Cache) Setex(key string, val interface{}, t time.Duration) {
	c.cache.set(key, val, t)
}

func (c *Cache) Get(key string) interface{} {
	val, ok := c.cache.get(key)
	if !ok {
		return nil
	}
	return val
}

func (c *Cache) Del(key string) {
	c.cache.del(key)
}

func (c *Cache) Incr(key string) error {
	return c.cache.incrby(key, 1)
}

func (c *Cache) Incrby(key string, s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	return c.cache.incrby(key, n)
}

func (c *Cache) Keys(exp string) ([]KV, error) {
	return c.cache.keys(exp)
}

func (c *Cache) Ttl(key string) (time.Duration, bool) {
	return c.cache.ttl(key)
}

func (c *Cache) Save(filename string) error {
	return c.cache.saveFile(filename)
}

func (c *Cache) Load(filename string) error {
	return c.cache.loadFile(filename)
}
