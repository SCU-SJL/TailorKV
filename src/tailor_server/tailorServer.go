package main

import (
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"tailor"
	"tailor_server/config"
	"tailor_server/handler"
	"time"
)

var (
	maxSizeOfDatagram int
	defaultExpiration time.Duration
	cleanCycle        time.Duration
	asyncCleanCycle   time.Duration
	concurrency       uint8
	savingPath        string
)

func main() {
	// read configuration
	conf := config.GetConfig(".." + string(os.PathSeparator) + ".." + string(os.PathSeparator) +
		"resource" + string(os.PathSeparator) + "config.xml")
	resolveConfig(*conf)

	// start tailor
	cache := tailor.NewCache(defaultExpiration, cleanCycle, asyncCleanCycle, concurrency, nil)

	// start server
	listener, err := net.Listen("tcp", "localhost:8448")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			cache.Save(savingPath, nil)
			log.Fatal(err)
		}
		go handler.HandleConn(conn, cache, savingPath, maxSizeOfDatagram)
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

	savingPath = conf.SavingPath + conf.FileName
}

func parseStr(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return i
}
