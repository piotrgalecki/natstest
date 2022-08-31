package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	HubCount = 500
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("Failed to %s: %s", msg, err)
	} else {
		log.Println("Succeeded to ", msg)
	}
}

func main() {
	nc, err := nats.Connect("nats.dev.siden.io", nats.Name("Scale test publisher"))
	failOnError(err, "connect to NATS")
	defer nc.Close()

	for {
		// sent msg to all hubs
		for hub := 1; hub <= HubCount; hub++ {
			mboxName := "hub-" + strconv.Itoa(hub)
			err := nc.Publish(mboxName, []byte("All is Well"))
			if err != nil {
				fmt.Println("failed to publish to ", mboxName, err)
			}
		}

		fmt.Println("Sent messages to all ", HubCount, " hubs")

		time.Sleep(60 * time.Second)
	}
}
