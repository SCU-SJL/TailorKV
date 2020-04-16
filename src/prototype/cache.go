package prototype

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
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

// TODO monitor is required
type watcher struct {
}

type cache struct {
	defaultExpiration      time.Duration
	defaultClearExpiration time.Duration
	items                  map[string]Item
	mu                     sync.RWMutex
	afterDel               func(string, interface{})
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
	case uint32:
		item.Data = item.Data.(uint32) + uint32(n)
	case float32:
		item.Data = item.Data.(float32) + float32(n)
	case float64:
		item.Data = item.Data.(float64) + float64(n)
	default:
		return fmt.Errorf("cannot incre the value of %s", key)
	}
	c.items[key] = item
	return nil
}

func (c *cache) incrfby(key string, n float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, found := c.find(key)
	if !found {
		return fmt.Errorf("item %s does not exist", key)
	}
	switch item.Data.(type) {
	case float32:
		item.Data = item.Data.(float32) + float32(n)
	case float64:
		item.Data = item.Data.(float64) + n
	default:
		fmt.Errorf("value of %s is not a float", key)
	}
	c.items[key] = item
	return nil
}

func (c *cache) del(key string) {
	c.mu.Lock()
	val, hasHandler := c.doDel(key)
	c.mu.Unlock()
	if hasHandler {
		c.afterDel(key, val)
	}
}

func (c *cache) doDel(key string) (interface{}, bool) {
	if c.afterDel != nil {
		if item, found := c.find(key); found {
			delete(c.items, key)
			return item.Data, true
		}
	}
	delete(c.items, key)
	return nil, false
}

type kv struct {
	key string
	val interface{}
}

// delExpired can be called by user or cache itself
func (c *cache) delExpired() {
	var itemsWithHandler []kv
	c.mu.Lock()
	for k, v := range c.items {
		if v.Expired() {
			val, hasHandler := c.doDel(k)
			if hasHandler {
				itemsWithHandler = append(itemsWithHandler, kv{k, val})
			}
		}
	}
	c.mu.Unlock()
	for _, item := range itemsWithHandler {
		c.afterDel(item.key, item.val)
	}
}

func (c *cache) saveFile(filename string) error {
	file, err := os.Create(filename)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return err
	}
	err = c.save(file)
	if err != nil {
		return err
	}
	return nil
}

func (c *cache) save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("error registering item types with gob")
		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, item := range c.items {
		gob.Register(item.Data)
	}
	err = enc.Encode(&c.items)
	return
}

func (c *cache) loadFile(filename string) error {
	file, err := os.Open(filename)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return err
	}
	err = c.load(file)
	if err != nil {
		return err
	}
	return nil
}

func (c *cache) load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, v := range items {
			_, found := c.find(k)
			if !found {
				c.items[k] = v
			}
		}
	}
	return err
}
