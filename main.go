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
	"github.com/go-redis/redis"
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

	// goripr wrapper, used all the time
	rdb, err = goripr.NewClient(goripr.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabase,
	})
	if err != nil {
		log.Fatalln("Could not establish redis database connection:", err)
	}
	// Redis client, used for initialization purposed only.
	initRdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabase,
	})
	defer initRdb.Close()

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
	log.Println("Initialized folder structure")

	if err = updateRedisDatabase(initRdb, rdb, cfg); err != nil {
		log.Fatalln(err)
	}
	log.Println("Initialized redis database content")
}

func main() {
	defer publisher.Close()
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
