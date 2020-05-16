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
	cls
	exit
	quit
)

var errType = []string{"Success", "SyntaxErr", "NotFound", "Existed",
	"NeSaveFailed", "ExSaveFailed", "LoadFailed"}

type Command struct {
	op  string
	key string
	val string
	exp string
}

func HandleConn(conn net.Conn, ipAddr *string) {
	ok, err := authorized(conn)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Print("Enter the password: ")
		var password string
		_, err = fmt.Scanln(&password)
		for err != nil {
			fmt.Println("invalid input")
			fmt.Print("Enter again: ")
			_, err = fmt.Scanln(&password)
		}
		_, err = conn.Write([]byte(password))
		if err != nil {
			log.Fatal(err)
		}
		resp := make([]byte, 1)
		_, err = conn.Read(resp)
		if err != nil {
			log.Fatal(err)
		}
		if resp[0] != 0 {
			fmt.Println("Wrong password")
			return
		}
	}

	for {
		fmt.Print(*ipAddr + ":8448-->:")
		command, err := readCommand()
		if err != nil {
			fmt.Println(err)
			fmt.Println()
			continue
		}
		switch command.op {
		case "set":
			handleCommandWithOneParam(conn, set, command)
		case "setex":
			handleCommandWithOneParam(conn, setex, command)
		case "setnx":
			handleCommandWithOneParam(conn, setnx, command)
		case "get":
			res, err := handleGet(conn, command)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(res)
		case "del":
			handleCommandWithOneParam(conn, del, command)
		case "unlink":
			handleCommandWithOneParam(conn, unlink, command)
		case "incr":
			handleCommandWithOneParam(conn, incr, command)
		case "incrby":
			handleCommandWithOneParam(conn, incrby, command)
		case "ttl":
			res, err := handleTtl(conn, command)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(res)
		case "keys":
			err := handleKeys(conn, command)
			if err != nil {
				fmt.Println(err)
			}
		case "cnt":
			res, err := handleCnt(conn, command)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(res)
		case "save":
			err := handleSave(conn, command)
			if err != nil {
				fmt.Println(err)
			}
		case "load":
			err := handleCommandWithNoResp(conn, load, command, true)
			if err != nil {
				fmt.Println(err)
			}
		case "cls":
			err := handleCommandWithNoResp(conn, cls, command, true)
			if err != nil {
				fmt.Println(err)
			}
		case "exit", "quit":
			_ = handleCommandWithNoResp(conn, exit, command, false)
			return
		}
	}
}

func handleGet(conn net.Conn, command *Command) (string, error) {
	sendDatagram(conn, get, command)
	msg := make([]byte, 1)
	_, err := conn.Read(msg)
	if err != nil {
		return "", err
	}
	if msg[0] != 0 {
		return "", errors.New(errType[msg[0]])
	}
	val := make([]byte, 4096)
	n, err := conn.Read(val)
	if err != nil {
		return "", err
	}
	return string(val[:n]), nil
}

func handleTtl(conn net.Conn, command *Command) (string, error) {
	sendDatagram(conn, ttl, command)
	msg := make([]byte, 1)
	_, err := conn.Read(msg)
	if err != nil {
		return "", err
	}
	if msg[0] != 0 {
		fmt.Print(errType[msg[0]])
	}
	val := make([]byte, 128)
	n, err := conn.Read(val)
	if err != nil {
		return "", err
	}
	return string(val[:n]), nil
}

func handleKeys(conn net.Conn, command *Command) error {
	sendDatagram(conn, keys, command)
	msg := make([]byte, 1)
	_, err := conn.Read(msg)
	if err != nil {
		return err
	}
	if msg[0] != 0 {
		err = printErrMsg(conn)
		return err
	}

	buf := make([]byte, 1024*1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	arr, err := protocol.GetKeys(buf[:n])
	if err != nil {
		return err
	}
	for _, k := range arr {
		fmt.Println(k)
	}
	return nil
}

func handleCnt(conn net.Conn, command *Command) (string, error) {
	sendDatagram(conn, cnt, command)
	msg := make([]byte, 1)
	_, err := conn.Read(msg)
	if err != nil {
		return "", err
	}
	count := make([]byte, 64)
	n, err := conn.Read(count)
	if err != nil {
		return "", err
	}
	return string(count[:n]), nil
}

func handleSave(conn net.Conn, command *Command) error {
	sendDatagram(conn, save, command)
	fmt.Print("NeCache: ")
	err := printErrMsg(conn)
	if err != nil {
		return err
	}
	fmt.Print("ExCache: ")
	err = printErrMsg(conn)
	return err
}

func handleCommandWithNoResp(conn net.Conn, op byte, command *Command, printErr bool) error {
	sendDatagram(conn, op, command)
	if printErr {
		err := printErrMsg(conn)
		return err
	}
	return nil
}

func handleCommandWithOneParam(conn net.Conn, op byte, command *Command) {
	sendDatagram(conn, op, command)
	err := printErrMsg(conn)
	if err != nil {
		fmt.Println(err)
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

func printErrMsg(conn net.Conn) error {
	errMsg := make([]byte, 64)
	n, err := conn.Read(errMsg)
	if err != nil {
		return err
	}
	if n == 1 {
		fmt.Println(errType[errMsg[0]])
	} else {
		fmt.Printf("errMsg: %s\n", errMsg[:n])
	}
	return nil
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

	if length > 1 && paramArr[1] == "-h" {
		printUsage(paramArr[0])
		return nil, errors.New("check TailorKV document for more info")
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
	case "set", "setex", "setnx", "auth",
		"get", "del", "unlink", "incr", "incrby",
		"ttl", "keys", "cnt", "save", "load", "cls", "exit", "quit":
		return nil
	default:
		return errors.New("illegal command: " + op)
	}
}

func checkCommand(op string, size int) error {
	lenErr := errors.New("wrong number of params")
	switch op {
	case "cnt", "cls", "exit", "quit":
		if size != 0 {
			return lenErr
		}
	case "get", "del", "unlink", "incr", "ttl", "keys", "auth":
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
	case "save", "load":
		if size > 1 {
			return lenErr
		}
	}
	return nil
}

func printUsage(op string) {
	fmt.Print("USAGE: ")
	switch op {
	case "auth":
		fmt.Println("auth [password]")
	case "cnt", "cls", "exit", "quit":
		fmt.Printf("%s\n", op)
	case "get", "del", "unlink", "incr", "ttl":
		fmt.Printf("%s [key]\n", op)
	case "keys":
		fmt.Println("keys [regular expression]")
	case "set", "setnx":
		fmt.Printf("%s [key] [val]\n", op)
	case "incrby":
		fmt.Println("incrby [key] [addition(Integer)]")
	case "setex":
		fmt.Println("setex [key] [val] [expiration]")
	case "save", "load":
		fmt.Printf("\n%s ## use default filepath\n", op)
		fmt.Printf("%s [filename]  ## use the given filename(doesn't change the Dir)\n", op)
	}
}

func authorized(conn net.Conn) (bool, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		return false, err
	}
	if buf[0] == 0 {
		return true, nil
	}
	return false, nil
}
