package handler

import (
	"bufio"
	"errors"
	"fmt"
	"log"
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

var errType = []string{"Success", "SyntaxErr", "NotFound", "Existed",
	"NeSaveFailed", "ExSaveFailed", "LoadFailed"}

type Command struct {
	op  string
	key string
	val string
	exp string
}

func HandleConn(conn net.Conn) {
	for {
		fmt.Print("localhost:8448-->:")
		command, err := readCommand()
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch command.op {
		case "set":
			sendDatagram(conn, set, command)
			printErrMsg(conn)
		case "setex":
			sendDatagram(conn, setex, command)
			printErrMsg(conn)
		case "setnx":
			sendDatagram(conn, setnx, command)
			printErrMsg(conn)
		case "get":
			sendDatagram(conn, get, command)
			msg := make([]byte, 1)
			_, err := conn.Read(msg)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if msg[0] != 0 {
				fmt.Println(errType[msg[0]])
				continue
			}
			val := make([]byte, 4096)
			n, err := conn.Read(val)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(val[:n]))
		case "del":
			sendDatagram(conn, del, command)
			printErrMsg(conn)
		case "unlink":
			sendDatagram(conn, unlink, command)
			printErrMsg(conn)
		case "incr":
			sendDatagram(conn, incr, command)
			printErrMsg(conn)
		case "incrby":
			sendDatagram(conn, incrby, command)
			printErrMsg(conn)
		case "ttl":
			sendDatagram(conn, ttl, command)
			msg := make([]byte, 1)
			_, err := conn.Read(msg)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if msg[0] != 0 {
				fmt.Print(errType[msg[0]])
			}
			val := make([]byte, 128)
			n, err := conn.Read(val)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(val[:n]))
		case "cnt":
			sendDatagram(conn, cnt, command)
			msg := make([]byte, 1)
			_, err := conn.Read(msg)
			if err != nil {
				fmt.Println(err)
				continue
			}
			count := make([]byte, 64)
			n, err := conn.Read(count)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(count[:n]))
		case "save":
			sendDatagram(conn, save, command)
			fmt.Print("NeCache: ")
			printErrMsg(conn)
			fmt.Print("ExCache: ")
			printErrMsg(conn)
		case "load":
			sendDatagram(conn, load, command)
			printErrMsg(conn)
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
	_, err := conn.Write(datagram)
	if err != nil {
		log.Fatal(err)
	}
}

func printErrMsg(conn net.Conn) {
	errMsg := make([]byte, 64)
	n, err := conn.Read(errMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	if n == 1 {
		fmt.Println(errType[errMsg[0]])
	} else {
		fmt.Printf("errMsg: %s\n", errMsg[:n])
	}
}

func readCommand() (*Command, error) {
	in := bufio.NewReader(os.Stdin)
	input, err := in.ReadString('\n')
	input = strings.Replace(input, "\r\n", "", -1)
	input = strings.Replace(input, "\n", "", -1)
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
		err = checkCommand(paramArr[0], 0)
		if err != nil {
			return nil, err
		}
		command.op = paramArr[0]
	} else if length == 2 {
		err = checkCommand(paramArr[0], 1)
		if err != nil {
			return nil, err
		}
		command.op = paramArr[0]
		command.key = paramArr[1]
	} else if length == 3 {
		err = checkCommand(paramArr[0], 2)
		if err != nil {
			return nil, err
		}
		command.op = paramArr[0]
		command.key = paramArr[1]
		command.val = paramArr[2]
	} else if length == 4 {
		err = checkCommand(paramArr[0], 3)
		if err != nil {
			return nil, err
		}
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

func checkCommand(op string, size int) error {
	lenErr := errors.New("wrong number of params")
	switch op {
	case "cnt", "save", "load":
		if size != 0 {
			return lenErr
		}
	case "get", "del", "unlink", "incr", "ttl", "keys":
		if size != 1 {
			return lenErr
		}
	case "set", "setnx", "incrby":
		if size != 2 {
			return lenErr
		}
	case "setex":
		if size != 3 {
			return lenErr
		}
	}
	return nil
}
