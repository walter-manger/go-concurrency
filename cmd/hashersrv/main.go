package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/walter-manger/go-concurrency/pkg/api"
)

var signalChan chan os.Signal

func init() {
	signalChan = make(chan os.Signal)
}

func shutdown(hapi *api.HasherAPI) {
	if hapi == nil {
		return
	}
	<-signalChan
	close(hapi.RequestChannel)
	log.Println("Hasher Service Shutting Down Gracefully")
	hapi.HasherJobs().Wait()
}

func main() {

	log.Printf("Starting Hasher Service...\n\n")

	port := flag.String("addr", "8080", "The port to listen on")
	flag.Parse()

	hapi := api.NewHasherAPI(*port)
	hapi.Start()

	signal.Notify(signalChan, os.Interrupt)

	shutdown(hapi)
}
