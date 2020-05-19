package main

import (
	"flag"
	"log"
	"net"
	"tailor_client/handler"
)

var ipAddr = flag.String("ip", "localhost", "ip address of host")
var port = flag.String("p", "8448", "port number")

func main() {
	flag.Parse()
	tcpAddr, err := net.ResolveTCPAddr("tcp4", *ipAddr+":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	handler.HandleConn(conn, ipAddr)
}
