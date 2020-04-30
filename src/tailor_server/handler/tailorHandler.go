package handler

import (
	"net"
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

func HandleConn(conn net.Conn, cache *tailor.Cache, savingPath string) {
	defer cache.Save(savingPath, nil)
	over := false
	for !over {
		op, err := readCommand(conn)
		if err != nil {
			break
		}
		switch op {
		case setex:
			doSetex(conn, cache)
		}
	}
}

func readCommand(conn net.Conn) (byte, error) {
	op := make([]byte, 1)
	_, err := conn.Read(op)
	return op[0], err
}
