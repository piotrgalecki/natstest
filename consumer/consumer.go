package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func logEvent(hub int, err error, msg string) {
	if err != nil {
		log.Println("[Hub ", hub, "] Failed to ", msg, "error", err)
	} else {
		log.Println("[Hub ", hub, "] Succeeded to", msg)
	}
}

func consumer(hub int, natsServer string) {
	hubMbox := "hub-" + strconv.Itoa(hub)

	// connect to NATS server
	var nc *nats.Conn
	var err error
	for {
		nc, err = nats.Connect(natsServer, nats.Name(hubMbox))
		logEvent(hub, err, "connect to nats server")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer nc.Close()

	// create async subscription to hub-n topic
	for {
		_, err := nc.Subscribe(hubMbox, func(m *nats.Msg) {
			log.Printf("Hub %d received a message", hub)
		})
		logEvent(hub, err, "subscribe to mailbox")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	var forever chan struct{}
	<-forever
}

func main() {
	// get hubCount from environement
	hubCount := 100
	hubCountStr := os.Getenv("hubCount")
	if hubCountStr != "" {
		hubCount64, _ := strconv.ParseInt(hubCountStr, 0, 64)
		hubCount = int(hubCount64)
	}

	// get natsServer from environement
	natsServer := "localhost"
	natsServerStr := os.Getenv("natsServer")
	if natsServerStr != "" {
		natsServer = natsServerStr
	}

	// get replica number from environement
	replica := 0
	hostname := os.Getenv("HOSTNAME")
	ss := strings.Split(hostname, "statefulset-")
	if len(ss) > 1 {
		replica64, _ := strconv.ParseInt(ss[1], 0, 64)
		replica = int(replica64)
	}

	log.Println("NATS consumer test - hubCount", hubCount, "natsServer", natsServer, "replica", replica)

	for hub := replica*hubCount + 1; hub <= replica*hubCount+hubCount; hub++ {
		go consumer(hub, natsServer)
		time.Sleep(10 * time.Millisecond)
	}

	var forever chan struct{}
	<-forever
}
