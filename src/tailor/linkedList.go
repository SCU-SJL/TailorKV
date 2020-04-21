package tailor

import (
	"fmt"
	"sync"
)

type node struct {
	data interface{}
	next *node
	prev *node
}

type LinkedList struct {
	size int
	head *node
	tail *node
	mu   sync.RWMutex
}

func (list *LinkedList) Size() int {
	list.mu.RLock()
	defer list.mu.RUnlock()
	return list.size
}

func (list *LinkedList) Get(n int) (interface{}, error) {
	list.mu.RLock()
	defer list.mu.RUnlock()
	if err := list.illegalIndexCheck(n); err != nil {
		return nil, err
	}
	res := list.node(n)
	return res.data, nil
}

func (list *LinkedList) Set(n int, data interface{}) error {
	list.mu.Lock()
	defer list.mu.Unlock()
	if err := list.illegalIndexCheck(n); err != nil {
		return err
	}
	res := list.node(n)
	res.data = data
	return nil
}

func (list *LinkedList) GetFirst() (interface{}, error) {
	list.mu.RLock()
	defer list.mu.RUnlock()
	if err := list.illegalIndexCheck(0); err != nil {
		return nil, err
	}
	return list.head.data, nil
}

func (list *LinkedList) GetLast() (interface{}, error) {
	list.mu.RLock()
	defer list.mu.RUnlock()
	if err := list.illegalIndexCheck(list.size - 1); err != nil {
		return nil, err
	}
	return list.tail.data, nil
}

func (list *LinkedList) AddFirst(data interface{}) {
	cur := &node{
		data: data,
		prev: nil,
	}
	list.mu.Lock()
	defer list.mu.Unlock()
	if list.head != nil {
		cur.next = list.head
		list.head.prev = cur
	} else {
		list.tail = cur
	}
	list.head = cur
	list.size++
}

func (list *LinkedList) AddLast(data interface{}) {
	cur := &node{
		data: data,
		next: nil,
	}
	list.mu.Lock()
	defer list.mu.Unlock()
	if list.tail != nil {
		cur.prev = list.tail
		list.tail.next = cur
	} else {
		list.head = cur
	}
	list.tail = cur
	list.size++
}

func (list *LinkedList) RemoveFirst() (interface{}, error) {
	list.mu.Lock()
	defer list.mu.Unlock()
	if err := list.illegalIndexCheck(0); err != nil {
		return nil, err
	}
	res := list.head
	list.head = list.head.next
	if list.head != nil {
		list.head.prev = nil
	}
	list.size--
	return res.data, nil
}

func (list *LinkedList) RemoveLast() (interface{}, error) {
	list.mu.Lock()
	defer list.mu.Unlock()
	if err := list.illegalIndexCheck(list.size - 1); err != nil {
		return nil, err
	}
	res := list.tail
	list.tail = res.prev
	if list.tail != nil {
		list.tail.next = nil
	}
	list.size--
	return res.data, nil
}

func (list *LinkedList) Offer(data interface{}) {
	list.AddLast(data)
}

func (list *LinkedList) Poll() (interface{}, error) {
	return list.RemoveFirst()
}

func (list *LinkedList) Push(data interface{}) {
	list.AddLast(data)
}

func (list *LinkedList) Pop() (interface{}, error) {
	return list.RemoveLast()
}

func (list *LinkedList) node(n int) *node {
	var res *node
	if n < list.size>>1 {
		res = list.head
		for i := 0; i < n; i++ {
			res = res.next
		}
	} else {
		res = list.tail
		for i := list.size - 1; i > n; i-- {
			res = res.prev
		}
	}
	return res
}

func (list *LinkedList) illegalIndexCheck(n int) error {
	if list.size == 0 || n > list.size-1 {
		return fmt.Errorf("index %d out of bounds for length %d", n, list.size)
	}
	return nil
}
