package main

import (
	"log"
	"net"
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
	_, _ = conn.Write([]byte{0})
	_, _ = conn.Write([]byte("name sjl 5000"))
}
