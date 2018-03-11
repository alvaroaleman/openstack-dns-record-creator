package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

const queueName = "notifications.info"

func main() {
	amqpUser := os.Getenv("AMQP_USER")
	amqpPass := os.Getenv("AMQP_PASS")
	amqpHost := os.Getenv("AMQP_HOST")
	amqpConnectionString := fmt.Sprintf("amqp://%s:%s@%s/", amqpUser, amqpPass, amqpHost)

	amqpConnection, err := amqp.Dial(amqpConnectionString)
	defer amqpConnection.Close()
	if err != nil {
		log.Fatalf("Error creating amqp connection: '%s'", err)
	}

	amqpChannel, err := amqpConnection.Channel()
	defer amqpChannel.Close()
	if err != nil {
		log.Fatalf("Error creating amqp channel: '%s'", err)
	}

	amqpQueue, err := amqpChannel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error creating amqp queue: '%s'", err)
	}

	msgs, err := amqpChannel.Consume(
		amqpQueue.Name,
		"coredns-record-creator",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Error registering consumer: '%s'", err)
	}

	halt := make(chan bool)
	go handleMessage(msgs)
	<-halt
}

func handleMessage(msgChannel <-chan amqp.Delivery) {
	for msg := range msgChannel {
		log.Printf("Got message:\n'%s'\n", msg)
	}
}
