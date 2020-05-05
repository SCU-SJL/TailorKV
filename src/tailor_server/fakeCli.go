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
	testGet(conn, "age")

	//testSetex(conn)
	//testGet(conn)
	//
	//testSetnx(conn)
	//testGet(conn)

	//testUnlink(conn)
	//<-time.After(2*time.Second)
	//testGet(conn)

	testIncr(conn)
	testGet(conn, "age")
	testIncrby(conn)
	testGet(conn, "age")
}

func testSetex(conn net.Conn) {
	datagram := getDatagram(setex, "name", "Jack Ma", "5000")
	_, _ = conn.Write(datagram)
	printErrMsg("[setex]", conn)
}

func testGet(conn net.Conn, key string) {
	datagram := getDatagram(get, key, "", "")
	_, _ = conn.Write(datagram)
	errMsg := printErrMsg("[get]", conn)
	if errMsg == 0 {
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		fmt.Println("[get] name = ", string(buf[:n]))
	}
}

func testSet(conn net.Conn) {
	datagram := getDatagram(set, "age", "20", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[set]", conn)
}

func testSetnx(conn net.Conn) {
	datagram := getDatagram(setnx, "name", "Pony Ma", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[setnx]", conn)
}

func testDel(conn net.Conn) {
	datagram := getDatagram(del, "name", "", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[del]", conn)
}

func testUnlink(conn net.Conn) {
	datagram := getDatagram(unlink, "name", "", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[unlink]", conn)
}

func testIncr(conn net.Conn) {
	datagram := getDatagram(incr, "age", "", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[incr]", conn)
}

func testIncrby(conn net.Conn) {
	datagram := getDatagram(incrby, "age", "2", "")
	_, _ = conn.Write(datagram)
	printErrMsg("[incrby]", conn)
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
