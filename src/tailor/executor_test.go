package tailor

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"
)

var initMap = map[string]string{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e", "f": "f"}

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

func TestCache_Keys(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	for k, v := range initMap {
		c.Set(k, v)
	}
	c.Setex("ex", "ex", 3*time.Second)
	<-time.After(500 * time.Millisecond)
	res, _ := c.Keys("[A-z]+")
	if len(res) != 7 {
		t.Errorf("keys failed: len of cache = %d", len(res))
	}
}

func TestCache_Incrby(t *testing.T) {
	c := NewCache(NoExpiration, nil)

	// regular int test
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 1000; i++ {
		key := "Int" + string(i)
		num := rand.Intn(MaxInt/2 - 1)
		c.Set(key, num)
		err := c.Incrby(key, strconv.Itoa(num))
		if err != nil {
			t.Errorf("Incr %s with valye of %d failed: %v", key, num, err)
			return
		}
		after := c.Get(key).(int)
		if after != 2*num {
			t.Errorf("Incr wrong, expected:%d, actual:%d", 2*num, after)
		}
	}

	// border test
	c.Set("border", uint8(33))
	if err := c.Incrby("border", strconv.Itoa(256)); err == nil {
		t.Errorf("Incr didn't throw err")
	}
	v := c.Get("border")
	if int(v.(uint8)) != 33 {
		t.Errorf("Incr changed value, expected:33, actual:%d", v)
	}

	// ... More test
}

func TestCache_Save(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	for i := 0; i < 10; i++ {
		k := "k" + strconv.Itoa(i)
		v := "v" + strconv.Itoa(i)
		c.Set(k, v)
	}
	c.Setex("exp", "exp10s", 10*time.Second)
	ok := make(chan bool, 2)
	c.Save("testCopy", ok)
	if o := <-ok; !o {
		t.Errorf("Save neCache failed")
	}
	if o := <-ok; !o {
		t.Errorf("Save exCache failed")
	}
	c.Cls()
	err := c.Load("testCopy")
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}
	if c.Cnt() != 11 {
		t.Errorf("size of cache is not correct")
	}
	fmt.Println(c.Ttl("exp"))
}

func TestCache_AddDelHandler(t *testing.T) {
	c := NewCache(NoExpiration, nil)
	c.AddDelHandler(func(key string, val interface{}) {
		fmt.Println("delete ", key, " with value of ", val)
	})
	c.Setex("exp", "exp1s", 1*time.Second)
	<-time.After(2 * time.Second)
}

func TestCpuNum(t *testing.T) {
	fmt.Println(runtime.NumCPU())
}
