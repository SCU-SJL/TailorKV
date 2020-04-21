package tailor

import (
	"fmt"
	"math"
	"testing"
	"time"
)

var c = NewCache(NoExpiration, nil)

var keys = []string{"name", "id", "address", "bigInt"}
var vals = []interface{}{"sjl", "2017141461144", "Chengdu", math.MaxInt64}

func TestCache_DelExpired(t *testing.T) {
	c.AddDelHandler(func(key string, val interface{}) {
		fmt.Printf("delete key:%s val:%v\n", key, val)
	})
	c.Setex("name", "sjl", 1*time.Second)
	c.Setex("id", "2017141461144", 1*time.Second)
	c.Setex("bigInt", math.MaxInt64, 2*time.Second)
	c.Setex("smallInt", math.MinInt64, 2*time.Second)
	ticker := time.After(4 * time.Second)
	<-ticker
	printCache(c)
}

func printCache(c *Cache) {
	for k, v := range c.neCache.items {
		fmt.Println("key:", k, "val:", v)
	}
	fmt.Println("---------------------------")
	for k, v := range c.exCache.items {
		fmt.Println("key:", k, "val:", v)
	}
}

func TestCache_Unlink(t *testing.T) {
	c.AddDelHandler(func(key string, val interface{}) {
		fmt.Printf("delete key:%s val:%v\n", key, val)
	})
	for i := 0; i < len(keys); i++ {
		c.Set(keys[i], vals[i])
	}
	for _, key := range keys {
		c.Unlink(key)
	}
	ticker := time.After(5 * time.Second)
	<-ticker
	for _, key := range keys {
		v := c.Get(key)
		if v != nil {
			t.Errorf("%s is not deleted, val: %v", key, v)
		}
	}
	if c.neCache.asyncQueue.Size() > 0 {
		t.Errorf("async queue is not empty")
	}
}

func TestCache_Watching(t *testing.T) {
	op := func(c *Cache) {
		fmt.Println("daemon report: len = ", c.Cnt())
	}
	c.ReplaceDaemonOp(1*time.Second, op)
	_ = c.StartWatching()
	go func() {
		for i := 0; i < len(keys); i++ {
			c.Set(keys[i], vals[i])
		}
	}()
	go func() {
		c.Set("k11", 11)
		c.Set("k22", 22)
		c.Set("k33", 33)
	}()
	<-time.After(5 * time.Second)
	_ = c.StopWatchingSync()

	fmt.Println("stop for 2s")
	<-time.After(2 * time.Second)

	fmt.Println("start watching now")

	_ = c.StartWatching()
	_ = c.StartWatching()
	_ = c.StartWatching()
	_ = c.StartWatching()

	fmt.Println("replace op after 3s")
	<-time.After(3 * time.Second)
	go func() {
		op2 := func(c *Cache) {
			fmt.Println("new daemon op")
		}
		c.ReplaceDaemonOp(1*time.Second, op2)
		_ = c.StartWatching()
	}()

	<-time.After(5 * time.Second)
}

func TestCache_StartWatching(t *testing.T) {
	op := func(c *Cache) {
		fmt.Println("daemon report: len = ", c.Cnt())
	}
	c.ReplaceDaemonOp(1*time.Second, op)
	_ = c.StartWatching()
	fmt.Println(c.StartWatching(), " --1")
	fmt.Println(c.StartWatching(), " --2")
	fmt.Println(c.StartWatching(), " --3")
	fmt.Println(c.StartWatching(), " --4")
	<-time.After(3 * time.Second)
}
