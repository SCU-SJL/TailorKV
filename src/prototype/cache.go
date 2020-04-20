package prototype

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	// claim to use the defaultExpiration of neCache
	DefaultExpiration time.Duration = 0
	NoExpiration      time.Duration = -1
)

const (
	Uint = iota
	Uint8
	Uint16
	Uint32
	Uint64
	Int
	Int8
	Int16
	Int32
	Int64

	MaxUint   = ^uint(0)
	MaxUint8  = ^uint8(0)
	MaxUint16 = ^uint16(0)
	MaxUint32 = ^uint32(0)
	MaxUint64 = ^uint64(0)

	MaxInt   = int(^uint(0) >> 1)
	MinInt   = ^MaxInt
	MaxInt8  = int8(^uint8(0) >> 1)
	MinInt8  = ^MaxInt8
	MaxInt16 = int16(^uint16(0) >> 1)
	MinInt16 = ^MaxInt16
	MaxInt32 = int32(^uint32(0) >> 1)
	MinInt32 = ^MaxInt32
	MaxInt64 = int64(^uint64(0) >> 1)
	MinInt64 = ^MaxInt64
)

type Item struct {
	Data       interface{}
	Expiration int64
}

func (item Item) Expired() bool {
	return item.Expiration >= 0 &&
		time.Now().UnixNano() > item.Expiration
}

type cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	afterDel          func(string, interface{})
}

func newCache(de time.Duration, m map[string]Item) *cache {
	if de < 0 {
		de = NoExpiration
	}
	if m == nil {
		m = make(map[string]Item)
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func (c *cache) addDelHandler(f func(string, interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterDel = f
}

func (c *cache) set(key string, val interface{}, lastFor time.Duration) {
	var ex int64
	if lastFor == DefaultExpiration {
		if c.defaultExpiration > 0 {
			ex = time.Now().Add(c.defaultExpiration).UnixNano()
		} else {
			ex = -1
		}
	} else if lastFor > 0 {
		ex = time.Now().Add(lastFor).UnixNano()
	} else {
		ex = -1
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

func (c *cache) ttl(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.find(key)
	if !found {
		return time.Duration(0), false
	}
	return time.Unix(0, item.Expiration).Sub(time.Now()), true
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
		if overflowed(Int, int64(item.Data.(int)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(int) + int(n)
	case int8:
		if overflowed(Int8, int64(item.Data.(int8)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(int8) + int8(n)
	case int16:
		if overflowed(Int16, int64(item.Data.(int16)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(int16) + int16(n)
	case int32:
		if overflowed(Int32, int64(item.Data.(int32)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(int32) + int32(n)
	case int64:
		if overflowed(Int64, int64(item.Data.(int64)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(int64) + n
	case uint:
		if overflowed(Uint, int64(item.Data.(uint)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(uint) + uint(n)
	case uint8:
		if overflowed(Uint8, int64(item.Data.(uint8)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(uint8) + uint8(n)
	case uint16:
		if overflowed(Uint16, int64(item.Data.(uint16)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(uint16) + uint16(n)
	case uint32:
		if overflowed(Uint32, int64(item.Data.(uint32)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(uint32) + uint32(n)
	case uint64:
		if overflowed(Uint64, int64(item.Data.(uint64)), n) {
			return fmt.Errorf("incr overflowed")
		}
		item.Data = item.Data.(uint64) + uint64(n)
	case float32:
	case float64:
		return fmt.Errorf("cannot incr a float")
	default:
		return fmt.Errorf("cannot incre the value of %s", key)
	}
	c.items[key] = item
	return nil
}

func overflowed(tp int, left, right int64) bool {
	switch tp {
	case Uint:
		return (right > 0 && uint64(left) > uint64(MaxUint)-uint64(right)) || (right < 0)
	case Uint8:
		return (right > 0 && left > int64(MaxUint8)-right) || (right < 0)
	case Uint16:
		return (right > 0 && left > int64(MaxUint16)-right) || (right < 0)
	case Uint32:
		return (right > 0 && left > int64(MaxUint32)-right) || (right < 0)
	case Uint64:
		return (right > 0 && uint64(left) > MaxUint64-uint64(right)) || (right < 0)
	case Int:
		return (right > 0 && left > int64(MaxInt)-right) ||
			(right < 0 && left < int64(MinInt)-right)
	case Int8:
		return (right > 0 && left > int64(MaxInt8)-right) ||
			(right < 0 && left < int64(MinInt8)-right)
	case Int16:
		return (right > 0 && left > int64(MaxInt16)-right) ||
			(right < 0 && left < int64(MinInt16)-right)
	case Int32:
		return (right > 0 && left > int64(MaxInt32)-right) ||
			(right < 0 && left < int64(MinInt32)-right)
	case Int64:
		return (right > 0 && left > MaxInt64-right) ||
			(right < 0 && left < MinInt64-right)
	default:
		return true
	}
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
		if item, found := c.items[key]; found {
			delete(c.items, key)
			return item.Data, true
		}
	}
	delete(c.items, key)
	return nil, false
}

type KV struct {
	key string
	val interface{}
}

func (kv *KV) Key() string {
	return kv.key
}

func (kv *KV) Val() interface{} {
	return kv.val
}

func (c *cache) delExpired() {
	c.mu.Lock()
	var itemsWithHandler []KV
	ticker := time.After(100 * time.Millisecond)
loop:
	for {
		select {
		case <-ticker:
			break loop
		default:
			count := 0
			for k, v := range c.items {
				count++
				if v.Expired() {
					val, hasHandler := c.doDel(k)
					if hasHandler {
						itemsWithHandler = append(itemsWithHandler, KV{k, val})
					}
				}
				if count > 4 {
					break
				}
			}
		}
	}
	c.mu.Unlock()

	go func() {
		for _, item := range itemsWithHandler {
			c.afterDel(item.key, item.val)
		}
	}()
}

func (c *cache) saveFile(filename string) error {
	file, err := os.Create(filename)
	defer func() {
		if file != nil {
			_ = file.Close()
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
			_ = file.Close()
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

func (c *cache) keys(exp string) ([]KV, error) {
	reg, err := regexp.Compile(exp)
	if err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make([]KV, 0)
	for k, v := range c.items {
		if reg.Match([]byte(k)) {
			res = append(res, KV{k, v})
		}
	}
	return res, nil
}

func (c *cache) cnt() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *cache) cls() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}
