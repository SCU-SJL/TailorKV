package handler

import (
	"encoding/json"
	"net"
	"protocol"
	"tailor"
)

const (
	setex byte = iota
	setnx
	set
	get
	del
	unlink
	incr
	incrby
	ttl
	keys
	cnt
)

func HandleConn(conn net.Conn, cache *tailor.Cache, savingPath string, maxSizeOfDatagram int) {
	defer cache.Save(savingPath, nil)
	for {
		datagram, err := readDatagram(conn, maxSizeOfDatagram)
		if err != nil {
			break
		}
		switch datagram.Op {
		case setex:
			doSetex(cache, datagram, conn)
		}
	}
}

func readDatagram(conn net.Conn, maxSize int) (*protocol.Protocol, error) {
	buf := make([]byte, maxSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var datagram protocol.Protocol
	err = json.Unmarshal(buf[:n], &datagram)
	if err != nil {
		return nil, err
	}
	return &datagram, nil
}
