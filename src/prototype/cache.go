package prototype

import (
	"fmt"
	"sync"
	"time"
)

const (
	DefaultExpiration time.Duration = 0
	NoExpiration      time.Duration = -1
)

type Item struct {
	Data       interface{}
	Expiration int64
}

func (item Item) Expired() bool {
	return item.Expiration > 0 &&
		time.Now().UnixNano() > item.Expiration
}

type Cache struct {
	cache   *cache
	watcher *watcher
}

type watcher struct {
}

type cache struct {
	defaultExpiration      time.Duration
	defaultClearExpiration time.Duration
	items                  map[string]Item
	mu                     sync.RWMutex
	handleDel              func(string, interface{})
}

func (c *cache) set(key string, val interface{}, lastFor time.Duration) {
	var ex int64
	if lastFor == DefaultExpiration {
		lastFor = c.defaultExpiration
	}
	if lastFor > 0 {
		ex = time.Now().Add(lastFor).UnixNano()
	}
	c.mu.Lock()
	// it seems that 'defer' adds ~200ns (saw on github)
	defer c.mu.Unlock()
	c.items[key] = Item{
		Data:       val,
		Expiration: ex,
	}
}

func (c *cache) setnx(key string, val interface{}, lastFor time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.find(key)
	if found {
		return fmt.Errorf("item %s already exists", key)
	}
	c.set(key, val, lastFor)
	return nil
}

// find - find item from map
// thread-unsafe
func (c *cache) find(key string) (Item, bool) {
	item, found := c.items[key]
	if !found || item.Expired() {
		return Item{}, false
	}
	return item, true
}

func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.find(key)
	if !found {
		return nil, false
	}
	return item.Data, true
}

func (c *cache) ttl(key string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.find(key)
	if !found {
		return time.Time{}, false
	}
	return time.Unix(0, item.Expiration), true
}

func (c *cache) incrby(key string, n int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, found := c.find(key)
	if !found {
		return fmt.Errorf("item %s does not exist", key)
	}
	switch item.Data.(type) {
	case int:
		item.Data = item.Data.(int) + int(n)
	case int16:
		item.Data = item.Data.(int16) + int16(n)
	case int32:
		item.Data = item.Data.(int32) + int32(n)
	case int64:
		item.Data = item.Data.(int64) + n
	case uint:
		item.Data = item.Data.(uint) + uint(n)
	case uint16:
		item.Data = item.Data.(uint16) + uint16(n)
	}
	return nil
}
