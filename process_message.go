package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/Teeworlds-Server-Moderation/common/events"
	"github.com/Teeworlds-Server-Moderation/common/mqtt"
	"github.com/jxsl13/goripr"
)

func processMessage(msg mqtt.Message, rdb *goripr.Client, publisher *mqtt.Publisher, cfg *Config) error {

	switch msg.Topic {
	case events.TypePlayerJoined:
		event := events.NewPlayerJoinedEvent()
		err := event.Unmarshal(msg.Payload)
		if err != nil {
			return fmt.Errorf("unable to unmarshal PlayerJoinedEvent: %s", err)
		}
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
		return fmt.Errorf("processMessage: unexpected topic: %s", msg.Topic)
	}
	return nil
}
