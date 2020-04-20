package prototype

import (
	"fmt"
	"testing"
	"time"
)

var c = NewCache(NoExpiration, nil)

func TestCache_AddDelHandler(t *testing.T) {
	f := func(key string, val interface{}) {
		fmt.Println("del ", key, " with value of ", val)
	}
	c.AddDelHandler(f)
	c.Set("name", "SJL")
	c.Setex("id", "2017141461144", 5*time.Second)
	c.Del("name")
	c.Del("id")
	if c.Get("name") != nil || c.Get("id") != nil {
		t.Errorf("item name or id is not deleted")
	}
}

func TestCache_Incr(t *testing.T) {
	c.Set("num1", 1) //int
	c.Set("num2", int16(2))
	c.Set("num3", int8(3))
	c.Set("num4", int32(4))
	c.Set("num5", int64(5))
	c.Set("num6", uint(123))
	c.Set("num7", uint8(12))
	c.Set("num8", uint16(123))
	c.Set("num9", uint32(123))
	c.Set("num10", uint64(123))
	if err := c.Incr("num1"); err != nil {
		t.Errorf("num1 failed: %v", err)
	}
	if err := c.Incr("num2"); err != nil {
		t.Errorf("num2 failed: %v", err)
	}
	if err := c.Incr("num3"); err != nil {
		t.Errorf("num3 failed: %v", err)
	}
	if err := c.Incr("num4"); err != nil {
		t.Errorf("num4 failed: %v", err)
	}
	if err := c.Incr("num5"); err != nil {
		t.Errorf("num5 failed: %v", err)
	}
	if err := c.Incr("num6"); err != nil {
		t.Errorf("num6 failed: %v", err)
	}
	if err := c.Incr("num7"); err != nil {
		t.Errorf("num7 failed: %v", err)
	}
	if err := c.Incr("num8"); err != nil {
		t.Errorf("num8 failed: %v", err)
	}
	if err := c.Incr("num9"); err != nil {
		t.Errorf("num9 failed: %v", err)
	}
	if err := c.Incr("num10"); err != nil {
		t.Errorf("num10 failed: %v", err)
	}
}

func TestCache_Incr_Border(t *testing.T) {
	keys := []string{"MinInt", "MinInt8", "MinInt16", "MinInt32", "MinInt64",
		"MaxUint", "MaxUint8", "MaxUint16", "MaxUint32", "MaxUint64"}
	vals := []interface{}{MinInt, MinInt8, MinInt16, MinInt32, MinInt64,
		MaxUint, MaxUint8, MaxUint16, MaxUint32, MaxUint64}
	for i := range keys {
		c.Set(keys[i], vals[i])
	}
	for i := 0; i < 5; i++ {
		if err := c.Incrby(keys[i], "-1"); err == nil {
			t.Errorf("key[%s] failed", keys[i])
		}
	}
	for i := 5; i < 10; i++ {
		if err := c.Incr(keys[i]); err == nil {
			t.Errorf("key[%s] failed", keys[i])
		}
	}
	//Kvs, err := c.Keys("[A-z]*")
	//if err != nil {
	//	t.Errorf("keys failedL %v", err)
	//}
	//for _, kv := range Kvs {
	//	fmt.Printf("key: %s val: %v\n", kv.Key(), kv.Val())
	//}
}

func TestCache_Ttl(t *testing.T) {
	c.Setex("name", "sjl", 200*time.Millisecond)
	time.Sleep(1 * time.Second)
	_, exist := c.Ttl("name")
	if exist {
		t.Errorf("ttl failed, a key is not expired")
	}

	c.Setex("id", "2017141461144", 50*time.Second)
	time.Sleep(1 * time.Second)
	ttl, exist := c.Ttl("id")
	if !exist {
		t.Errorf("ttl failed, a key is expired advancedly")
	}
	if ttl > 49*time.Second || ttl < 48*time.Second {
		t.Errorf("ttl failed, expected: 48~49s, actual:%v", ttl)
	}
}

func TestCache_Cleaner(t *testing.T) {
	keys := []string{"t500ms", "t1000s", "t1500ms", "t2000ms"}
	c.AddDelHandler(func(key string, val interface{}) {
		fmt.Println("key:", key, "val:", val, " expired!")
	})
	for i, k := range keys {
		n := 500 * i
		c.Setex(k, k, time.Duration(n)*time.Millisecond)
	}
	wait := time.After(3 * time.Second)
	<-wait
	if len(c.cache.items) == 4 {
		t.Errorf("daemon cleaner went wrong")
	}
}

func TestCache_ReplaceDaemonOp(t *testing.T) {
	c.Set("name", "shaojiale")
	c.Set("id", "2017141461144")
	f := func(c *Cache) {
		fmt.Println("name = ", c.Get("name"))
	}
	c.ReplaceDaemonOp(1*time.Second, f)
	ticker := time.After(5 * time.Second)
	<-ticker
	f = func(c *Cache) {
		fmt.Println("id = ", c.Get("id"))
	}
	c.ReplaceDaemonOp(1*time.Second, f)
	ticker = time.After(5 * time.Second)
	<-ticker
}
