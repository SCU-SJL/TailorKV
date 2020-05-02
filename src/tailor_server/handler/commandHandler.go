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

func doGet(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val, found := cache.Get(key)
	if !found {
		_, _ = conn.Write([]byte{NotFound})
		return
	}
	_, _ = conn.Write([]byte{Success})
	_, _ = conn.Write([]byte(val.(string)))
}
