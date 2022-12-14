package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	STREAM_NAME = "teststream"
)

func logEvent(hub int, err error, msg string) {
	if err != nil {
		log.Println("[Hub ", hub, "] Failed to ", msg, "error", err)
	} else {
		log.Println("[Hub ", hub, "] Succeeded to", msg)
	}
}

func consumer(hub int, natsServer string, useJetstream bool, jsName string) {
	hubMbox := "testmbox-" + strconv.Itoa(hub)

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

	var js nats.JetStreamContext
	if useJetstream {
		// Create JetStream Context
		js, err = nc.JetStream(nats.PublishAsyncMaxPending(256))
		logEvent(hub, err, "create jetstream context")

		// add jetstream consumer
		_, err = js.AddConsumer(jsName,
			&nats.ConsumerConfig{
				Durable:       hubMbox,
				Description:   hubMbox + " consumer",
				DeliverPolicy: nats.DeliverNewPolicy, // only deliver new messages
				AckPolicy:     nats.AckExplicitPolicy,
				FilterSubject: jsName + "." + hubMbox + ".>",
			})
		logEvent(hub, err, "create jetstream consumer: stream "+jsName)
	}

	// create sync subscription to the hub-n topic
	var sub *nats.Subscription
	for {
		subject := hubMbox + ".>"
		if useJetstream {
			subject = jsName + "." + subject
			sub, err = js.PullSubscribe(subject, hubMbox)
		} else {
			sub, err = nc.SubscribeSync(subject)
		}
		logEvent(hub, err, "create subscription to "+subject)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	// receive messages
	for {
		if useJetstream {
			msgs, _ := sub.Fetch(10)
			for _, m := range msgs {
				m.Ack() // for js use explicit ack
				log.Println("[Hub ", hub, "] Received js message w/ subject", m.Subject, "data", string(m.Data))
			}
		} else {
			m, err := sub.NextMsg(1 * time.Minute)
			if err != nil {
				continue
			}
			log.Println("[Hub ", hub, "] Received message w/ subject", m.Subject, "data", string(m.Data))
		}
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

	jsCount := 1
	jsCountStr := os.Getenv("JS_COUNT")
	if jsCountStr != "" {
		jsCount64, _ := strconv.ParseInt(jsCountStr, 0, 64)
		jsCount = int(jsCount64)
	}

	// get replica number from environement
	replica := 0
	hostname := os.Getenv("HOSTNAME")
	ss := strings.Split(hostname, "statefulset-")
	if len(ss) > 1 {
		replica64, _ := strconv.ParseInt(ss[1], 0, 64)
		replica = int(replica64)
	}

	log.Println("NATS consumer test - hubCount", hubCount, "natsServer", natsServer, "js", useJetstream, "jsCount", jsCount, "replica", replica)

	if useJetstream {
		// connect to NATS server
		var nc *nats.Conn
		var err error
		for {
			nc, err = nats.Connect(natsServer)
			logEvent(0, err, "connect to nats server")
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		defer nc.Close()

		// Create JetStream Context
		js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
		logEvent(0, err, "create jetstream context")

		// Create JestStream for messages to hubs
		for strm := 1; strm <= jsCount; strm++ {
			jsName := STREAM_NAME + "-" + strconv.Itoa(strm)

			_, err = js.AddStream(&nats.StreamConfig{
				Name:              jsName,                  // stream name
				Description:       jsName,                  // description
				Subjects:          []string{jsName + ".>"}, // stream subjects
				Retention:         nats.LimitsPolicy,       // messages are retained until limit is reached
				MaxConsumers:      100,                     // max_consumers
				MaxMsgs:           8000,                    // max_msgs
				MaxBytes:          0,                       // max_bytes
				Discard:           nats.DiscardOld,         // discard policy
				MaxAge:            10 * time.Minute,        // max_age
				MaxMsgsPerSubject: 0,                       // max_msgs_per_subject
				MaxMsgSize:        1024,                    // max_msg_size
				Storage:           nats.MemoryStorage,      // storage type
				Replicas:          2,                       // num_replicas
				//NoAck             bool            `json:"no_ack,omitempty"`
				//Template          string          `json:"template_owner,omitempty"`
				//Duplicates        time.Duration   `json:"duplicate_window,omitempty"`
				//Placement         *Placement      `json:"placement,omitempty"`
				//Mirror            *StreamSource   `json:"mirror,omitempty"`
				//Sources           []*StreamSource `json:"sources,omitempty"`
				//Sealed            bool            `json:"sealed,omitempty"`
				//DenyDelete        bool            `json:"deny_delete,omitempty"`
				//DenyPurge         bool            `json:"deny_purge,omitempty"`
				//AllowRollup       bool            `json:"allow_rollup_hdrs,omitempty"`
				//RePublish *SubjectMapping `json:"republish,omitempty // Allow republish of the message after being sequenced and stored
			})
			logEvent(0, err, "created hub jetstream"+jsName)
		}
	}

	// start hub simulators
	for hub := replica*hubCount + 1; hub <= replica*hubCount+hubCount; hub++ {
		jsName := STREAM_NAME + "-" + strconv.Itoa((hub-1)%jsCount+1)
		go consumer(hub, natsServer, useJetstream, jsName)
		time.Sleep(10 * time.Millisecond)
	}

	var forever chan struct{}
	<-forever
}
