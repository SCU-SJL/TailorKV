package main

import (
	"fmt"
	"log"
	"net"
	"protocol"
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
	data := &protocol.Protocol{
		Op:  0,
		Key: "name",
		Val: "Jack Ma",
		Exp: "500",
	}
	datagram, _ := data.GetJsonBytes()
	_, _ = conn.Write(datagram)
	errMsg := make([]byte, 1)
	_, _ = conn.Read(errMsg)
	fmt.Println(errMsg[0])
}
