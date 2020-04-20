package tailor

import (
	"fmt"
	"math"
	"testing"
	"time"
)

var c = NewCache(NoExpiration, nil)

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
