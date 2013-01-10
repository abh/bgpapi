package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)

	log.New(os.Stderr, "bgpapi", log.LstdFlags)
	log.SetPrefix("bgpapi ")

	log.Println("Starting")

	go bgpReader()
	httpServer()

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt)

	<-terminate
	log.Printf("signal received, stopping")
}
