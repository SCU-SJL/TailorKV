package main

import (
	"fmt"
	"log"
	"net"
	"protocol"
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

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:8448")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	//testSetex(conn)
	testGet(conn)
}

func testSetex(conn net.Conn) {

	data := &protocol.Protocol{
		Op:  setex,
		Key: "name",
		Val: "Jack Ma",
		Exp: "5000",
	}
	datagram, _ := data.GetJsonBytes()
	_, _ = conn.Write(datagram)
	errMsg := make([]byte, 1)
	_, _ = conn.Read(errMsg)
	fmt.Println("[setex] errMsg =", errMsg[0])
}

func testGet(conn net.Conn) {
	data := &protocol.Protocol{
		Op:  get,
		Key: "name",
		Val: "",
		Exp: "",
	}
	datagram, _ := data.GetJsonBytes()
	_, _ = conn.Write(datagram)
	errMsg := make([]byte, 1)
	_, _ = conn.Read(errMsg)
	fmt.Println("[get] errMsg = ", errMsg[0])
	if errMsg[0] == 0 {
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		fmt.Println("[get] name = ", string(buf[:n]))
	}
}
