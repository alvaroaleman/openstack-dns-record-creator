package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	amqpUser := os.Getenv("AMQP_USER")
	amqpPass := os.Getenv("AMQP_PASS")
	amqpHost := os.Getenv("AMQP_HOST")
	amqpConnectionString := fmt.Sprintf("amqp://%s:%s@%s/", amqpUser, amqpPass, amqpHost)
	log.Printf(amqpConnectionString)
}
