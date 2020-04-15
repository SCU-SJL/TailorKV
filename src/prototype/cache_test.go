package prototype

import (
	"fmt"
	"testing"
	"time"
)

var c = cache{items: map[string]Item{}, defaultExpiration: NoExpiration}

func TestCache_Find(t *testing.T) {
	c.set("name", "sjl", 10*time.Millisecond)
	time.Sleep(1 * time.Second)
	_, ok := c.find("name")
	if ok != false {
		t.Errorf("Func find() failed, expected = %v, actual = %v", false, ok)
	}
}

func TestCache_Ttl(t *testing.T) {
	fmt.Println("cur: ", time.Now())
	c.set("ttl", "5s", 5*time.Second)
	time.Sleep(5 * time.Second)
	ex, _ := c.ttl("ttl")
	fmt.Println(ex)
}
