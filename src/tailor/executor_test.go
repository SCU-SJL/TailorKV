package tailor

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestExecutor(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	c.set("name", "sjl")
	exec := newExecutor(c, 5)
	go exec.server()
	j := &job{
		op:   get,
		key:  "name",
		res:  response{},
		done: make(chan struct{}),
	}
	exec.execute(j)
	<-j.done
	fmt.Println(j.res.value)

	c.setex("id", "2017141461144", 2*time.Second)
	time.Sleep(500 * time.Millisecond)
	j = &job{
		op:   ttl,
		key:  "id",
		done: make(chan struct{}),
		res:  response{},
	}
	exec.execute(j)
	<-j.done
	fmt.Println(j.res.value)
}

func TestCache_Get(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	c.AddDelHandler(func(key string, val interface{}) {
		fmt.Println("del ", key, " val", val)
	})
	c.Set("name", "zhh")
	c.Set("and", "love")
	c.Set("name2", "sjl")
	c.Setex("ex", "ex2s", 1*time.Second)
	name := c.Get("name").(string)
	and := c.Get("and").(string)
	name2 := c.Get("name2").(string)
	if name != "zhh" || and != "love" || name2 != "sjl" {
		t.Errorf("Get failed, expected: 'zhh', actual:%v, expected: 'love'm, actual: %v, expected: 'sjl', actual: %v", name, and, name2)
	}
	<-time.After(2 * time.Second)
}

func TestCache_ReplaceDaemonOp(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	c.ReplaceDaemonOp(1*time.Second, func(c *Cache) {
		fmt.Println("daemon tick!")
	})
	err := c.StartWatching()
	if err != nil {
		t.Errorf("cannot start daemon watcher: %v", err)
	}
	<-time.After(3 * time.Second)
}

func TestCpuNum(t *testing.T) {
	fmt.Println(runtime.NumCPU())
}
