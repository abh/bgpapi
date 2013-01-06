package main

import (
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(4)

	go bgpReader()
	httpServer()
}
