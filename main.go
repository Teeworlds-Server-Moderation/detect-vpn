package main

// Connect to the broker, subscribe, and write messages received to a file

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Teeworlds-Server-Moderation/common/amqp"
	"github.com/Teeworlds-Server-Moderation/common/env"
	"github.com/Teeworlds-Server-Moderation/common/events"
	"github.com/Teeworlds-Server-Moderation/common/topics"
	"github.com/go-redis/redis"
	"github.com/jxsl13/goripr"
)

const (
	applicationID = "detect-vpn"
)

var (
	cfg        = &Config{}
	rdb        *goripr.Client
	subscriber *amqp.Subscriber
	publisher  *amqp.Publisher
)

func brokerCredentials(c *Config) (address, username, password string) {
	return c.BrokerAddress, c.BrokerUsername, c.BrokerPassword
}

// ExchangeCreator can be publisher or subscriber
type ExchangeCreator interface {
	CreateExchange(string) error
}

// QueueCreateBinder creates queues and binds them to exchanges
type QueueCreateBinder interface {
	CreateQueue(queue string) error
	BindQueue(queue, exchange string) error
}

func createExchanges(ec ExchangeCreator, exchanges ...string) {
	for _, exchange := range exchanges {
		if err := ec.CreateExchange(exchange); err != nil {
			log.Fatalf("Failed to create exchange '%s': %v\n", exchange, err)
		}
	}
}

func createQueueAndBindToExchanges(qcb QueueCreateBinder, queue string, exchanges ...string) {
	if err := qcb.CreateQueue(queue); err != nil {
		log.Fatalf("Failed to create queue '%s'\n", queue)
	}

	for _, exchange := range exchanges {
		if err := qcb.BindQueue(queue, exchange); err != nil {
			log.Fatalf("Failed to bind queue '%s' to exchange '%s'\n", queue, exchange)
		}

	}
}

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

	subscriber, err = amqp.NewSubscriber(brokerCredentials(cfg))
	if err != nil {
		log.Fatalln("Could not establish subscriber connection: ", err)
	}

	publisher, err = amqp.NewPublisher(brokerCredentials(cfg))
	if err != nil {
		log.Fatalln("Could not establish publisher connection: ", err)
	}

	createExchanges(
		publisher,
		topics.Broadcast,
	)

	createExchanges(
		subscriber,
		events.TypePlayerJoined,
	)

	createQueueAndBindToExchanges(
		subscriber,
		applicationID,
		events.TypePlayerJoined,
	)

	if err = initFolderStructure(cfg); err != nil {
		log.Fatalln(err)
	}
	log.Println("Initialized folder structure")

	// First one is used to update modified states of the file structure
	// second one is used to add IP ranges to the database
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
		next, err := subscriber.Consume(applicationID)
		if err != nil {
			log.Fatalln(err)
		}
		for msg := range next {
			if err := processMessage(string(msg.Body), rdb, publisher, cfg); err != nil {
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
