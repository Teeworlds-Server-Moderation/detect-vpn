package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/Teeworlds-Server-Moderation/common/amqp"
	"github.com/Teeworlds-Server-Moderation/common/events"
	"github.com/jxsl13/goripr"
)

func processMessage(msg string, rdb *goripr.Client, publisher *amqp.Publisher, cfg *Config) error {

	log.Printf("Processing Message: %s\n", msg)

	// the baseevent contains all necessary type information
	baseEvent := events.BaseEvent{}
	err := baseEvent.Unmarshal(msg)
	if err != nil {
		return err
	}

	switch baseEvent.Type {
	case events.TypePlayerJoined:
		event := events.NewPlayerJoinedEvent()
		err := event.Unmarshal(msg)
		if err != nil {
			return fmt.Errorf("unable to unmarshal PlayerJoinedEvent: %s", err)
		}
		log.Printf("Trying to find: '%s'\n", event.IP)
		reason, err := rdb.Find(event.IP)
		if errors.Is(goripr.ErrIPNotFound, err) {
			log.Printf("[NO VPN]: %s\n", event.IP)
			return nil
		} else if err != nil {
			return fmt.Errorf("unexpected error occurred: %s", err)
		}
		if err := requestBan(publisher, cfg, event.Player, reason, event.EventSource); err != nil {
			return err
		}
		log.Printf("[IS VPN]: %s\n", event.IP)
	default:
		return fmt.Errorf("processMessage: unexpected topic: %s", baseEvent.Type)
	}
	return nil
}
