package handler

import (
	"fmt"
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
	save
	load
	cls
	exit
	quit
)

type AESLogin struct {
	AuthRequired bool
	AuthPassword string
	AESKey       string
	AuthPassed   bool
}

func HandleConn(conn net.Conn, cache *tailor.Cache, savingDir, defaultSavingPath string, maxSizeOfDatagram int, login *AESLogin) {
	defer conn.Close()
	if loginErr := auth(conn, login); loginErr != nil {
		_, _ = conn.Write([]byte{1})
		return
	}
	_, err := conn.Write([]byte{0})
	if err != nil {
		return
	}

	defer func() {
		kvs, _ := cache.Keys("[A-z]+")
		for i := range kvs {
			fmt.Printf("key: %s, val: %v\n", kvs[i].Key(), kvs[i].Val())
		}
	}()

	for {
		datagram, err := readDatagram(conn, maxSizeOfDatagram)
		if err != nil {
			break
		}

		switch datagram.Op {
		case setex:
			doSetex(cache, datagram, conn)
		case setnx:
			doSetnx(cache, datagram, conn)
		case set:
			doSet(cache, datagram, conn)
		case get:
			doGet(cache, datagram, conn)
		case del:
			doDel(cache, datagram, conn)
		case unlink:
			doUnlink(cache, datagram, conn)
		case incr:
			doIncr(cache, datagram, conn)
		case incrby:
			doIncrby(cache, datagram, conn)
		case ttl:
			doTtl(cache, datagram, conn)
		case keys:
			doKeys(cache, datagram, conn)
		case cnt:
			doCnt(cache, conn)
		case save:
			doSave(savingDir, datagram, defaultSavingPath, cache, conn)
		case load:
			doLoad(savingDir, datagram, defaultSavingPath, cache, conn)
		case cls:
			doCls(cache, conn)
		case exit, quit:
			return
		}
	}
}

func readDatagram(conn net.Conn, maxSize int) (*protocol.Protocol, error) {
	buf := make([]byte, maxSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	datagram, err := protocol.GetDatagram(buf[:n])
	if err != nil {
		return nil, err
	}
	return datagram, nil
}
