package prototype

import (
	"fmt"
	"testing"
	"time"
)

var cache_ = NewCache(NoExpiration, nil)

func TestCache_AddDelHandler(t *testing.T) {
	f := func(key string, val interface{}) {
		fmt.Println("del ", key, " with value of ", val)
	}
	cache_.AddDelHandler(f)
	cache_.Set("name", "SJL")
	cache_.Setex("id", "2017141461144", 5*time.Second)
	cache_.Del("name")
	cache_.Del("id")
	if cache_.Get("name") != nil || cache_.Get("id") != nil {
		t.Errorf("item name or id is not deleted")
	}
}

func TestCache_Incr(t *testing.T) {
	cache_.Set("num1", 1)
	cache_.Set("num2", int16(2))
	cache_.Set("num3", int8(3))
	cache_.Set("num4", int32(4))
	cache_.Set("num5", int64(5))
	cache_.Set("num6", int8(0))
	if err := cache_.Incr("num1"); err != nil {
		t.Errorf("num1 failed: %v", err)
	}
	if err := cache_.Incr("num2"); err != nil {
		t.Errorf("num2 failed: %v", err)
	}
	if err := cache_.Incr("num3"); err != nil {
		t.Errorf("num3 failed: %v", err)
	}
	if err := cache_.Incr("num4"); err != nil {
		t.Errorf("num4 failed: %v", err)
	}
	if err := cache_.Incr("num5"); err != nil {
		t.Errorf("num5 failed: %v", err)
	}
	if err := cache_.Incrby("num6", -10000); err != nil {
		t.Errorf("num6 failed: %v", err)
	}
	fmt.Println(cache_.Get("num1"), cache_.Get("num2"), cache_.Get("num3"),
		cache_.Get("num4"), cache_.Get("num5"), cache_.Get("num6"))
	fmt.Printf("%T", cache_.Get("num6"))
}

func TestCache_Ttl(t *testing.T) {
	cache_.Setex("name", "sjl", 2*time.Second)
	time.Sleep(2 * time.Second)
	fmt.Println(cache_.Ttl("name"))
}
