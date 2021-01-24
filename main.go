package main

// Connect to the broker, subscribe, and write messages received to a file

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Teeworlds-Server-Moderation/common/mqtt"
)

var (
	topic         = "topic1"
	serverAddress = "tcp://localhost:1883"
	clientID      = "detect-vpn"
)

func init() {
	if brokerAddress := os.Getenv("BROKER_ADDRESS"); brokerAddress != "" {
		serverAddress = brokerAddress
	}
	if id := os.Getenv("BROKER_CLIENT_ID"); id != "" {
		clientID = id
	}

	if t := os.Getenv("BROKER_TOPIC"); t != "" {
		topic = t
	}

	log.Println("Initialized with address: ", serverAddress, " clientID: ", clientID)
}

func main() {

	subscriber, err := mqtt.NewSubscriber(serverAddress, clientID, topic)
	if err != nil {
		log.Fatalln("Could not establish subscriber connection: ", err)
	}
	defer subscriber.Close()
	go func() {
		for msg := range subscriber.Next() {
			log.Println("Received message: ", msg)
		}
	}()

	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Connection is up, press Ctrl-C to shutdown")
	<-sig
	fmt.Println("signal caught - exiting")
	fmt.Println("shutdown complete")
}
