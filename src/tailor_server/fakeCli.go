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

var errType = []string{"Success", "SyntaxErr", "NotFound", "Existed"}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:8448")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	testSet(conn)
	testGet(conn)

	testSetex(conn)
	testGet(conn)

	testSetnx(conn)
	testGet(conn)
}

func testSetex(conn net.Conn) {
	datagram := getDatagram(setex, "name", "Jack Ma", "5000")
	_, _ = conn.Write(datagram)
	printErrMsg("[setex]", conn)
}

func testGet(conn net.Conn) {
	datagram := getDatagram(get, "name", "", "")
	_, _ = conn.Write(datagram)
	errMsg := printErrMsg("[get]", conn)
	if errMsg == 0 {
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		fmt.Println("[get] name = ", string(buf[:n]))
	}
}

func testSet(conn net.Conn) {
	datagram := getDatagram(set, "name", "Jimmy S", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[set]", conn)
}

func testSetnx(conn net.Conn) {
	datagram := getDatagram(setnx, "name", "Pony Ma", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[setnx]", conn)
}

func getDatagram(op byte, key, val, exp string) []byte {
	data := &protocol.Protocol{
		Op:  op,
		Key: key,
		Val: val,
		Exp: exp,
	}
	datagram, _ := data.GetJsonBytes()
	return datagram
}

func printErrMsg(opName string, conn net.Conn) byte {
	errMsg := make([]byte, 1)
	_, _ = conn.Read(errMsg)
	fmt.Printf("%s errMsg = %s\n", opName, errType[errMsg[0]])
	return errMsg[0]
}
