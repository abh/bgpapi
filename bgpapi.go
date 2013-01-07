package main

import (
	"log"
	"os"
	"runtime"
)

func main() {
	// TODO(abh) need mutexes on the neighbors struct
	runtime.GOMAXPROCS(1)

	log.New(os.Stderr, "bgpapi", log.LstdFlags)
	log.SetPrefix("bgpapi")

	go bgpReader()
	httpServer()
}
