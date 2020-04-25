package tailor

import (
	"fmt"
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
