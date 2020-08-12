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
	executor *executor
}

func NewCache(defaultExpiration, cleanCycle, unlinkCycle time.Duration, concurrency uint8, m map[string]Item) *Cache {
	// if expiry time is not greater than zero, then make it NoExpiration
	var nec, exc *cache
	if defaultExpiration <= 0 {
		nec = newCache(NoExpiration, unlinkCycle, m)
		exc = newCache(NoExpiration, unlinkCycle, m)
	} else {
		exc = newCache(defaultExpiration, unlinkCycle, m)
		nec = exc
	}

	// clean expired data twice each second
	cl := defaultCleaner(cleanCycle)
	C := &Cache{
		neCache:  nec,
		exCache:  exc,
		cleaner:  cl,
		wStopped: true,
	}

	// create a new executor
	exec := newExecutor(C, concurrency)
	C.executor = exec

	// start the daemon cleaner
	go cl.run(exc)
	// start the executor
	go exec.server()
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
	c.neCache.set(key, val, DefaultExpiration)
	if c.exCache != c.neCache {
		c.exCache.del(key)
	}
}

func (c *Cache) setnx(key string, val interface{}) bool {
	_, found := c.exCache.get(key)
	if found {
		return false
	}
	ok := c.neCache.setnx(key, val, DefaultExpiration)
	return ok
}

func (c *Cache) setex(key string, val interface{}, t time.Duration) {
	c.exCache.set(key, val, t)
	if c.neCache != c.exCache {
		c.neCache.del(key)
	}
}

func (c *Cache) get(key string) (interface{}, bool) {
	val, ok := c.neCache.get(key)
	if ok {
		return val, true
	}
	val, ok = c.exCache.get(key)
	if ok {
		return val, true
	}
	return nil, false
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
	info := c.neCache.sIncrby(key, 1)
	if info == 2 {
		return incrErr(key, info)
	}
	if info == 1 {
		if c.exCache != c.neCache {
			info = c.exCache.sIncrby(key, 1)
			return incrErr(key, info)
		}
	}
	return nil
}

func (c *Cache) incrby(key string, s string) error {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("addition '%s' cannot be parsed to int64", s)
	}
	info := c.neCache.sIncrby(key, n)
	if info == 2 {
		return incrErr(key, info)
	}
	if info == 1 {
		if c.exCache != c.neCache {
			info = c.exCache.sIncrby(key, n)
			return incrErr(key, info)
		}
	}
	return nil
}

func incrErr(key string, info byte) error {
	switch info {
	case 1:
		return fmt.Errorf("key '%s' does not exist", key)
	case 2:
		return fmt.Errorf("value of '%s' cannot be parsed to int64", key)
	default:
		return nil
	}
}

func (c *Cache) ttl(key string) (time.Duration, bool) {
	return c.exCache.ttl(key)
}

/*
 * functions above is for Cache itself only.
 * functions below is exposed to users.
 */

// Keys get all KVs which match the regular expression.
// The result may not contain the KV which was set into
// the cache recently, as the Set operation is async.
func (c *Cache) Keys(exp string) ([]KV, error) {
	var res []KV
	res, err := c.neCache.keys(exp)
	if err != nil {
		return nil, err
	}
	if c.exCache != c.neCache {
		exKV, err := c.exCache.keys(exp)
		if err != nil {
			return nil, err
		}
		res = append(res, exKV...)
	}
	return res, nil
}

// Save param ok must be a chan with length of 2
func (c *Cache) Save(filename string, ok chan bool) {
	go func() {
		err := c.neCache.saveFile(filename + "ne")
		if err != nil {
			ok <- false
		}
		ok <- true
	}()
	if c.exCache != c.neCache {
		go func() {
			err := c.exCache.saveFile(filename + "ex")
			if err != nil {
				ok <- false
			}
			ok <- true
		}()
	} else {
		ok <- true
	}
}

func (c *Cache) Load(filename string) error {
	err := c.neCache.loadFile(filename + "ne")
	if err != nil {
		return err
	}
	if c.exCache != c.neCache {
		err = c.exCache.loadFile(filename + "ex")
		return err
	}
	return nil
}

func (c *Cache) Cls() {
	c.exCache.cls()
	if c.neCache != c.exCache {
		c.neCache.cls()
	}
}

func (c *Cache) Cnt() int {
	// the result contains the expired items
	// which are not cleaned when the func is called.
	cnt := c.neCache.cnt()
	if c.exCache != c.neCache {
		cnt += c.exCache.cnt()
	}
	return cnt
}

func (c *Cache) Set(key string, val interface{}) {
	newJob := &job{
		op:  set,
		key: key,
		val: val,
	}
	c.executor.execute(newJob)
}

func (c *Cache) Setnx(key string, val interface{}) bool {
	newJob := &job{
		op:   setnx,
		key:  key,
		val:  val,
		done: make(chan struct{}),
		res:  response{},
	}
	c.executor.execute(newJob)
	<-newJob.done
	return newJob.res.value.(bool)
}

func (c *Cache) Setex(key string, val interface{}, exp time.Duration) {
	newJob := &job{
		op:  setex,
		key: key,
		val: val,
		exp: exp,
	}
	c.executor.execute(newJob)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	newJob := &job{
		op:   get,
		key:  key,
		done: make(chan struct{}),
		res:  response{},
	}
	c.executor.execute(newJob)
	<-newJob.done
	return newJob.res.value, newJob.res.ok
}

func (c *Cache) Del(key string) {
	newJob := &job{
		op:  del,
		key: key,
	}
	c.executor.execute(newJob)
}

func (c *Cache) Unlink(key string) {
	newJob := &job{
		op:  unlink,
		key: key,
	}
	c.executor.execute(newJob)
}

func (c *Cache) Incr(key string) error {
	newJob := &job{
		op:   incr,
		key:  key,
		done: make(chan struct{}),
		res:  response{},
	}
	c.executor.execute(newJob)
	<-newJob.done
	return newJob.res.err
}

func (c *Cache) Incrby(key, addition string) error {
	newJob := &job{
		op:   incrby,
		key:  key,
		val:  addition,
		done: make(chan struct{}),
		res:  response{},
	}
	c.executor.execute(newJob)
	<-newJob.done
	return newJob.res.err
}

func (c *Cache) Ttl(key string) (time.Duration, bool) {
	newJob := &job{
		op:   ttl,
		key:  key,
		done: make(chan struct{}),
		res:  response{},
	}
	c.executor.execute(newJob)
	<-newJob.done
	return newJob.res.value.(time.Duration), newJob.res.ok
}
