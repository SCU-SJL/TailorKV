package handler

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"tailor"
	"time"
)

func readParam(conn net.Conn, n int) ([]string, error) {
	in := make([]byte, 64)
	nBytes, err := conn.Read(in)
	if err != nil {
		return nil, err
	}
	params := string(in[:nBytes])
	fmt.Println(params)
	res := strings.Split(params, " ")
	if len(res) != n {
		return nil, fmt.Errorf("syntax error, expected %d params, actual: %d params", n, len(res))
	}
	return res, err
}

func doSetex(conn net.Conn, cache *tailor.Cache) {
	params, err := readParam(conn, 3)
	if err != nil {
		log.Fatal(err)
		// ...
	}
	fmt.Println(params)
	key := params[0]
	val := params[1]
	exp, err := strconv.ParseInt(params[2], 10, 64)
	if err != nil {
		fmt.Println(err)
		// ...
	}
	cache.Setex(key, val, time.Duration(exp)*time.Millisecond)
}
