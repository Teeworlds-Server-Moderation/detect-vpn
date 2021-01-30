package main

// Connect to the broker, subscribe, and write messages received to a file

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Teeworlds-Server-Moderation/common/env"
	"github.com/Teeworlds-Server-Moderation/common/events"
	"github.com/Teeworlds-Server-Moderation/common/mqtt"
	"github.com/jxsl13/goripr"
)

var (
	clientID   = "detect-vpn"
	cfg        = &Config{}
	rdb        *goripr.Client
	subscriber *mqtt.Subscriber
	publisher  *mqtt.Publisher
)

func init() {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	rdb, err = goripr.NewClient(goripr.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword, // no password set
		DB:       cfg.RedisDatabase, // use default DB
	})
	if err != nil {
		log.Fatalln("Could not establish redis database connection:", err)
	}

	subscriber, err = mqtt.NewSubscriber(cfg.BrokerAddress, clientID, events.TypePlayerJoined)
	if err != nil {
		log.Fatalln("Could not establish subscriber connection: ", err)
	}

	publisher, err = mqtt.NewPublisher(cfg.BrokerAddress, clientID, "")
	if err != nil {
		log.Fatalln("Could not establish publisher connection: ", err)
	}

	if err = initFolderStructure(cfg); err != nil {
		log.Fatalln(err)
	}

	if err = updateRedisDatabase(rdb, cfg); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	defer subscriber.Close()
	defer rdb.Close()
	go func() {
		for msg := range subscriber.Next() {
			if err := processMessage(msg, rdb, publisher, cfg); err != nil {
				log.Printf("Error processing message: %s\n", err)
			}
		}
	}()

	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Connection is up, press Ctrl-C to shutdown")
	<-sig
	fmt.Println("Signal caught - exiting")
	fmt.Println("Shutdown complete")
}
