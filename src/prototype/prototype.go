package prototype

import (
	"strconv"
	"time"
)

type Cache struct {
	cache   *cache
	watcher *watcher
}

func NewCache(defaultExpiration time.Duration, m map[string]Item) *Cache {
	c := newCache(defaultExpiration, m)
	C := &Cache{
		cache: c,
	}
	return C
}

func (c *Cache) ReplaceDaemonOp(wi time.Duration, f func(*Cache)) {
	c.StopWatching()
	c.watcher.interval = wi
	c.watcher.op = f
}

func (c *Cache) StopWatching() {
	c.watcher.stop <- true
}

func (c *Cache) StartWatching(t time.Duration) {
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

func (c *Cache) DelExp() {
	c.cache.delExpired()
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