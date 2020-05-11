package handler

import (
	"net"
	"protocol"
	"strconv"
	"tailor"
	"time"
)

const (
	Success byte = iota
	SyntaxErr
	NotFound
	Existed
	NeSaveFailed
	ExSaveFailed
	LoadFailed
)

func doSetex(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val := datagram.Val
	exp, err := strconv.ParseInt(datagram.Exp, 10, 64)
	if err != nil {
		errMsg := []byte{SyntaxErr}
		_, _ = conn.Write(errMsg)
		return
	}
	cache.Setex(key, val, time.Duration(exp)*time.Millisecond)
	_, _ = conn.Write([]byte{Success})
}

func doSetnx(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val := datagram.Val
	ok := cache.Setnx(key, val)
	if ok {
		_, _ = conn.Write([]byte{Success})
		return
	}
	_, _ = conn.Write([]byte{Existed})
}

func doSet(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val := datagram.Val
	cache.Set(key, val)
	_, _ = conn.Write([]byte{Success})
}

func doGet(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val, found := cache.Get(key)
	if !found {
		_, _ = conn.Write([]byte{NotFound})
		return
	} else {
		_, err := conn.Write([]byte{Success})
		if err != nil {
			return
		}
	}
	_, _ = conn.Write([]byte(val.(string)))
}

func doDel(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	cache.Del(key)
	_, _ = conn.Write([]byte{Success})
}

func doUnlink(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	cache.Unlink(key)
	_, _ = conn.Write([]byte{Success})
}

func doIncr(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	err := cache.Incr(key)
	if err != nil {
		buf := []byte(err.Error())
		_, _ = conn.Write(buf)
		return
	}
	_, _ = conn.Write([]byte{Success})
}

func doIncrby(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val := datagram.Val
	err := cache.Incrby(key, val)
	if err != nil {
		buf := []byte(err.Error())
		_, _ = conn.Write(buf)
		return
	}
	_, _ = conn.Write([]byte{Success})
}

func doTtl(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	ttl, ok := cache.Ttl(key)
	if !ok {
		_, _ = conn.Write([]byte{NotFound})
	}
	_, _ = conn.Write([]byte{Success})
	_, _ = conn.Write([]byte(ttl.String()))
}

func doCnt(cache *tailor.Cache, conn net.Conn) {
	cnt := strconv.Itoa(cache.Cnt())
	_, _ = conn.Write([]byte{Success})
	_, _ = conn.Write([]byte(cnt))
}

func doSave(path string, cache *tailor.Cache, conn net.Conn) {
	status := make(chan bool, 2)
	cache.Save(path, status)

	if neOk := <-status; !neOk {
		_, _ = conn.Write([]byte{NeSaveFailed})
	} else {
		_, _ = conn.Write([]byte{Success})
	}

	if exOk := <-status; !exOk {
		_, _ = conn.Write([]byte{ExSaveFailed})
	} else {
		_, _ = conn.Write([]byte{Success})
	}
}

func doLoad(path string, cache *tailor.Cache, conn net.Conn) {
	err := cache.Load(path)
	if err != nil {
		_, _ = conn.Write([]byte{LoadFailed})
	} else {
		_, _ = conn.Write([]byte{Success})
	}
}

func doCls(cache *tailor.Cache, conn net.Conn) {
	cache.Cls()
	_, _ = conn.Write([]byte{Success})
}
