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
	save
	load
)

var errType = []string{"Success", "SyntaxErr", "NotFound", "Existed", "NeSaveFailed", "ExSaveFailed", "LoadFailed"}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "localhost:8448")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	//testSet(conn, "age", "abc")
	//testGet(conn, "age")

	//testSetex(conn)
	//testGet(conn)
	//
	//testSetnx(conn)
	//testGet(conn)

	//testUnlink(conn)
	//<-time.After(2*time.Second)
	//testGet(conn)

	//testIncr(conn, "age")
	//testSet(conn, "age", "23")
	//testIncrby(conn, "age", "23")
	//testGet(conn, "age")
	//
	//testSetnx(conn, "age", "55")
	//testGet(conn, "age")

	//testSetex(conn, "name", "Jack Ma")
	//testSetex(conn, "age", "20")
	//testIncrby(conn, "age", "5")
	//<-time.After(1 * time.Second)
	//testTtl(conn, "name")
	//testTtl(conn, "age")
	//testGet(conn, "age")
	//<-time.After(4 * time.Second)
	//testGet(conn, "name")
	//testGet(conn, "age")
	//<-time.After(10 * time.Second)
	//testCnt(conn)
	//testSave(conn)

	//testSet(conn, "ali", "Jack Ma")
	//testSet(conn, "tx", "Pony Ma")
	//testSave(conn)
	testCnt(conn)
	testLoad(conn)
	testGet(conn, "ali")
	testGet(conn, "tx")
}

func testSetex(conn net.Conn, key, val string) {
	sendDatagram(conn, setex, key, val, "5000")
	printErrMsg("[setex] "+key+"-"+val, conn)
}

func testGet(conn net.Conn, key string) {
	sendDatagram(conn, get, key, "", "")
	errMsg := printErrMsg("[get] "+key, conn)
	if errMsg == 0 {
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		fmt.Printf("[get] %s = %s\n", key, string(buf[:n]))
	}
}

func testSet(conn net.Conn, key, val string) {
	sendDatagram(conn, set, key, val, "")
	printErrMsg("[set] "+key+"-"+val, conn)
}

func testSetnx(conn net.Conn, key, val string) {
	sendDatagram(conn, setnx, key, val, "")
	printErrMsg("[setnx] "+key+"-"+val, conn)
}

func testDel(conn net.Conn, key string) {
	sendDatagram(conn, del, key, "", "")
	printErrMsg("[del] "+key, conn)
}

func testUnlink(conn net.Conn, key string) {
	sendDatagram(conn, unlink, key, "", "")
	printErrMsg("[unlink] "+key, conn)
}

func testIncr(conn net.Conn, key string) {
	sendDatagram(conn, incr, key, "", "")
	printErrMsg("[incr] "+key, conn)
}

func testIncrby(conn net.Conn, key, val string) {
	sendDatagram(conn, incrby, key, val, "")
	printErrMsg("[incrby] "+key+" with "+val, conn)
}

func testTtl(conn net.Conn, key string) {
	sendDatagram(conn, ttl, key, "", "")
	errMsg := printErrMsg("[ttl] "+key, conn)
	if errMsg == 0 {
		buf := make([]byte, 128)
		n, _ := conn.Read(buf)
		fmt.Println("[ttl]", key, string(buf[:n]))
	}
}

func testCnt(conn net.Conn) {
	sendDatagram(conn, cnt, "", "", "")
	printErrMsg("[cnt]", conn)
	buf := make([]byte, 16)
	n, _ := conn.Read(buf)
	fmt.Println("[cnt] = ", string(buf[:n]))
}

func testSave(conn net.Conn) {
	sendDatagram(conn, save, "", "", "")
	printErrMsg("[save] -neCache", conn)
	printErrMsg("[save] -exCache", conn)
}

func testLoad(conn net.Conn) {
	sendDatagram(conn, load, "", "", "")
	printErrMsg("[load]", conn)
}

func sendDatagram(conn net.Conn, op byte, key, val, exp string) {
	data := &protocol.Protocol{
		Op:  op,
		Key: key,
		Val: val,
		Exp: exp,
	}
	datagram, _ := data.GetJsonBytes()
	_, _ = conn.Write(datagram)
}

func printErrMsg(opName string, conn net.Conn) byte {
	errMsg := make([]byte, 64)
	n, _ := conn.Read(errMsg)
	if n == 1 {
		fmt.Printf("%s errMsg = %s\n", opName, errType[errMsg[0]])
	} else {
		fmt.Printf("%s errMsg = %s\n", opName, errMsg[:n])
	}
	return errMsg[0]
}
