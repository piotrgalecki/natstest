package main

import (
	"log"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

const HubCount = 500

func failOnError(client int, err error, msg string) {
	if err != nil {
		log.Panicf("Client ", client, "Failed to %s: %s", msg, err)
	} else {
		log.Println("Client ", client, "Succeeded to ", msg)
	}
}

func consumer(hub int) {
	// connect to NATS server
	nc, err := nats.Connect("nats.dev.siden.io")
	failOnError(hub, err, "connect to nats server")
	defer nc.Close()

	// create async subscription to hub-n topic
	subject := "hub-" + strconv.Itoa(hub)
	if _, err := nc.Subscribe(subject, func(m *nats.Msg) {
		log.Printf("Hub %d received a message", hub)
	}); err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}
	<-forever
}

func main() {

	for hub := 1; hub <= HubCount; hub++ {
		go consumer(hub)
		time.Sleep(10 * time.Millisecond)
	}

	var forever chan struct{}
	<-forever
}
