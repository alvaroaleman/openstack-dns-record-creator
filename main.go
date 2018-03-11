package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/streadway/amqp"
)

const queueName = "notifications.info"

func main() {
	amqpUser := os.Getenv("AMQP_USER")
	amqpPass := os.Getenv("AMQP_PASS")
	amqpHost := os.Getenv("AMQP_HOST")

	clientCert := os.Getenv("AMQP_CLIENCERTFILE")
	clientKey := os.Getenv("AMQP_CLIENTKEYFILE")
	caCertFilePath := os.Getenv("AMQP_CACERT")

	amqpScheme := "amqp"

	var tlsConfig *tls.Config
	if clientCert != "" && clientKey != "" && caCertFilePath != "" {
		amqpScheme = "amqps"
		log.Printf("Loading client cert and key..")
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			log.Fatalf("Error loading client cert and key: '%s'", err)
		}

		log.Printf("Loading cacert...")
		caCert, err := ioutil.ReadFile(caCertFilePath)
		if err != nil {
			log.Fatalf("Error reading cacert: '%s'", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   fmt.Sprintf("AMQP/%s", strings.Split(amqpHost, ":")[0]),
		}
	}

	amqpConnectionString := fmt.Sprintf("%s://%s:%s@%s/", amqpScheme, amqpUser, amqpPass, amqpHost)

	var amqpConnection *amqp.Connection
	var err error
	if amqpScheme == "amqp" {
		amqpConnection, err = amqp.Dial(amqpConnectionString)
	} else if amqpScheme == "amqps" {
		amqpConnection, err = amqp.DialTLS(amqpConnectionString, tlsConfig)
	}
	if err != nil {
		log.Fatalf("Error creating amqp connection: '%s'", err)
	}
	defer amqpConnection.Close()

	amqpChannel, err := amqpConnection.Channel()
	if err != nil {
		log.Fatalf("Error creating amqp channel: '%s'", err)
	}
	defer amqpChannel.Close()

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

	controller, err := newController()
	if err != nil {
		log.Fatalf("Error creating controller: '%v'", err)
	}

	halt := make(chan bool)
	log.Printf("Successfully finished startup!")
	go controller.handleMessage(msgs)
	<-halt
}

func (c *Controller) handleMessage(msgChannel <-chan amqp.Delivery) {
	for rawMsg := range msgChannel {
		var msg Message
		err := json.Unmarshal(rawMsg.Body, &msg)
		if err != nil {
			log.Printf("Error decoding message: %s", err)
			continue
		}
		if msg.EventType != "floatingip.update.end" {
			log.Printf("Ingoring event of type '%s'...", msg.EventType)
			continue
		}

		if msg.Payload.FloatingIP.FixedIPAddress == "" {
			go removeRecord(msg.Payload.FloatingIP.FloatingIPAddress)
			continue
		}

		go c.getInstanceNameAndCreateRecord(msg.Payload.FloatingIP.FloatingIPAddress)
	}
}

func (c *Controller) getInstanceNameAndCreateRecord(floatingIP string) {
	instanceName, err := c.getIntanceName(floatingIP)
	if err != nil {
		log.Printf("Error getting instance name for ip '%s': '%v'", floatingIP, err)
		return
	}

	err = createRecord(instanceName, floatingIP)
	if err != nil {
		log.Printf("Error creating record for intsance '%s': '%v'", instanceName, err)
	}
}

func (c *Controller) getIntanceName(floatingIP string) (string, error) {
	allPages, err := servers.List(c.OpenstackComputeClient, servers.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		return "", err
	}
	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return "", err
	}

	// addresses is []map[string]interface{} but we can not directly cast that...
	for _, server := range allServers {
		for _, val := range server.Addresses {
			if valList, ok := val.([]interface{}); ok {
				for _, valListItem := range valList {
					if valListItemMap, ok := valListItem.(map[string]interface{}); ok {
						if addressType, ok := valListItemMap["OS-EXT-IPS:type"]; ok && addressType == "floating" {
							if addr, ok := valListItemMap["addr"]; ok && addr == floatingIP {
								return server.Name, nil
							}
						}
					}
				}
			}
		}
	}

	return "", nil
}

func createRecord(name, address string) error {
	log.Printf("Not implemented: Creating record '%s' for address %s", name, address)
	return nil
}

func removeRecord(address string) error {
	// Kinda shitty to delete just based on address, however the FloatingIP events
	// after a deletion do not give any hints in regards to what the address was attached
	// to (both port_id and fixed_ip_address are null)
	// So the only alterantive would be to internally track that which is even worse
	log.Printf("Not implemented: Deleting record for address '%s'", address)
	return nil
}
