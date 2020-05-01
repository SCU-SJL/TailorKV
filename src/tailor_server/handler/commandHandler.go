package handler

import (
	"fmt"
	"net"
	"protocol"
	"strconv"
	"tailor"
	"time"
)

func doSetex(cache *tailor.Cache, datagram *protocol.Protocol, conn net.Conn) {
	key := datagram.Key
	val := datagram.Val
	exp, err := strconv.ParseInt(datagram.Exp, 10, 64)
	if err != nil {
		errMsg := []byte{1}
		_, _ = conn.Write(errMsg)
		return
	}
	cache.Setex(key, val, time.Duration(exp)*time.Millisecond)
	_, _ = conn.Write([]byte{0})
	fmt.Println(cache.Get(key))
}
