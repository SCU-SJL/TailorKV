package tailor

import (
	"fmt"
	"testing"
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
}
