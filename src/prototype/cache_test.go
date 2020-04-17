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

func TestCache_Save(t *testing.T) {
	c.set("name", "shaojiale", DefaultExpiration)
	c.set("id", "2017141461144", DefaultExpiration)
	err := c.saveFile("data.txt")
	if err != nil {
		t.Errorf("save failed: %v", err)
	}
}

func TestCache_Load(t *testing.T) {
	c.loadFile("data.txt")
	keys := []string{"name", "id"}
	for _, key := range keys {
		if _, ok := c.get(key); !ok {
			t.Errorf("key[%s] is not loaded", key)
		}
	}
	if _, ok := c.get("ttl"); ok {
		t.Errorf("key[%s] should be expired", keys)
	}
}

func TestCache_Keys(t *testing.T) {
	c.set("hello", "world", DefaultExpiration)
	c.set("name", "shaojiale", DefaultExpiration)
	res, err := c.keys("[lome]+")
	if err != nil {
		t.Errorf("test func keys failed: %v", err)
	}
	fmt.Println(res)
}
