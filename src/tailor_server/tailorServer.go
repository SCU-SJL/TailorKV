package main

import (
	"TailorKV/src/tailor"
	"TailorKV/src/tailor_server/config"
	"TailorKV/src/tailor_server/handler"
	"errors"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

var (
	maxSizeOfDatagram int
	defaultExpiration time.Duration
	cleanCycle        time.Duration
	asyncCleanCycle   time.Duration
	concurrency       uint8
	savingPath        string
	auth              bool
	password          string
	aesKey            string
	port              string
)

func main() {
	// read configuration
	conf := config.GetConfig(".." + string(os.PathSeparator) + "resource" + string(os.PathSeparator) + "config.xml")
	resolveConfig(*conf)
	login := &handler.AESLogin{
		AuthRequired: auth,
		AuthPassword: password,
		AESKey:       aesKey,
		AuthPassed:   false,
	}

	// start tailor
	cache := tailor.NewCache(defaultExpiration, cleanCycle, asyncCleanCycle, concurrency, nil)

	// start server
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			cache.Save(savingPath, nil)
		}
		go handler.HandleConn(conn, cache, conf.SavingDir, savingPath, maxSizeOfDatagram, login)
	}
}

func resolveConfig(conf config.TailorConfig) {
	maxSizeOfDatagram = int(parseStr(conf.MaxSizeofDatagram))

	i := parseStr(conf.DefaultExpiration)
	if i <= 0 {
		defaultExpiration = tailor.NoExpiration
	} else {
		defaultExpiration = time.Duration(i) * time.Millisecond
	}

	i = parseStr(conf.CleanCycle)
	if i <= 0 {
		log.Fatal("clean cycle must be greater than zero")
	}
	cleanCycle = time.Duration(i) * time.Millisecond

	i = parseStr(conf.AsyncCleanCycle)
	if i <= 0 {
		log.Fatal("async clean cycle must be greater than zero")
	}
	asyncCleanCycle = time.Duration(i) * time.Millisecond

	cc := conf.Concurrency
	if cc == "default" {
		concurrency = uint8(2 * runtime.NumCPU())
	} else {
		i, err := strconv.ParseUint(cc, 10, 8)
		if err != nil {
			log.Fatal(err)
		}
		concurrency = uint8(i)
	}

	if conf.Auth == "true" {
		auth = true
	} else if conf.Auth == "false" {
		auth = false
	} else {
		log.Fatal(errors.New("value of 'auth' in config.xml is invalid"))
	}
	port = conf.Port
	password = conf.Password
	aesKey = conf.AESKey
	savingPath = conf.SavingDir + conf.FileName
}

func parseStr(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return i
}
