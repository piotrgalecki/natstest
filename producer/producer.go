package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

func main() {
	// get hubCount from environement
	hubCount := 100
	hubCountStr := os.Getenv("HUB_COUNT")
	if hubCountStr != "" {
		hubCount64, _ := strconv.ParseInt(hubCountStr, 0, 64)
		hubCount = int(hubCount64)
	}

	// get natsServer from environement
	natsServer := "localhost"
	natsServerStr := os.Getenv("NATS_SERVER")
	if natsServerStr != "" {
		natsServer = natsServerStr
	}

	useJetstream := false
	jetStreamStr := os.Getenv("JETSTREAM")
	if jetStreamStr == "TRUE" || jetStreamStr == "true" {
		useJetstream = true
	}

	// get msgRate from environement
	msgRate := 1000
	msgRateStr := os.Getenv("MSG_RATE")
	if msgRateStr != "" {
		msgRate64, _ := strconv.ParseInt(msgRateStr, 0, 64)
		msgRate = int(msgRate64)
	}

	// get msgBurst from environement
	msgBurst := 100
	msgBurstStr := os.Getenv("MSG_BURST")
	if msgBurstStr != "" {
		msgBurst64, _ := strconv.ParseInt(msgBurstStr, 0, 64)
		msgBurst = int(msgBurst64)
	}

	sleepMs := 1000 / (msgRate / msgBurst)

	log.Println("NATS producer test - hubCount", hubCount, "natsServer", natsServer, "js", useJetstream,
		"msgRate", msgRate, "msgBurst", msgBurst, "sleepMs", sleepMs)

	// connect to NATS server
	var nc *nats.Conn
	var err error
	for {
		nc, err = nats.Connect(natsServer, nats.Name("NatsTestPublisher"))
		logEvent(0, err, "connect to nats server")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer nc.Close()

	var js nats.JetStreamContext
	if useJetstream {
		// Create JetStream Context
		js, err = nc.JetStream(nats.PublishAsyncMaxPending(256))
		logEvent(0, err, "create jetstream context")
	}

	hub := 1
	for {
		// sent msg burst
		for i := 0; i < msgBurst; i++ {
			mbox := "testmbox-" + strconv.Itoa(hub)
			if useJetstream {
				_, err = js.Publish("hubmsg."+mbox+".JsPerfTest", []byte("All is Well"))
			} else {
				err = nc.Publish(mbox+".PerfTest", []byte("All is Well"))
			}
			if err != nil {
				logEvent(hub, err, "publish message")
			}

			hub++
			if hub > hubCount {
				fmt.Println("Sent messages to all ", hubCount, " hubs")
				hub = 1
			}
		}

		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
}
