package handler

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"protocol"
	"strings"
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

type Command struct {
	op  string
	key string
	val string
	exp string
}

func HandleConn(conn net.Conn) {
	for {
		fmt.Println("localhost:8448-->:")
		command, err := readCommand()
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch command.op {
		case "set":
			sendDatagram(conn, set, command)
		case "setex":
			sendDatagram(conn, setex, command)
		}
	}
}

func sendDatagram(conn net.Conn, op byte, command *Command) {
	data := &protocol.Protocol{
		Op:  op,
		Key: command.key,
		Val: command.val,
		Exp: command.exp,
	}
	datagram, _ := data.GetJsonBytes()
	_, _ = conn.Write(datagram)
}

func readCommand() (*Command, error) {
	in := bufio.NewReader(os.Stdin)
	input, err := in.ReadString('\n')
	for err != nil {
		return nil, err
	}

	paramArr := strings.Split(input, " ")
	command := &Command{}
	length := len(paramArr)

	if length < 1 || length > 4 {
		return nil, errors.New("invalid input")
	}
	err = checkOp(paramArr[0])
	if err != nil {
		return nil, err
	}

	if length == 1 {
		command.op = paramArr[0]
	} else if length == 2 {
		command.op = paramArr[0]
		command.key = paramArr[1]
	} else if length == 3 {
		command.op = paramArr[0]
		command.key = paramArr[1]
		command.val = paramArr[2]
	} else if length == 4 {
		command.op = paramArr[0]
		command.key = paramArr[1]
		command.val = paramArr[2]
		command.exp = paramArr[3]
	}
	return command, nil
}

func checkOp(op string) error {
	switch op {
	case "set", "setex", "setnx",
		"get", "del", "unlink", "incr", "incrby",
		"ttl", "keys", "cnt", "save", "load":
		return nil
	default:
		return errors.New("illegal command: " + op)
	}
}
