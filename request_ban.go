package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/Teeworlds-Server-Moderation/common/dto"
	"github.com/Teeworlds-Server-Moderation/common/events"
	"github.com/Teeworlds-Server-Moderation/common/mqtt"
)

// serverTopic may be either the server's ip:port address or the broadcast topic
func requestBan(publisher *mqtt.Publisher, cfg *Config, player dto.Player, reason, source string) error {
	const ID = "detect-vpn"
	event := events.NewRequestCommandExecEvent()
	event.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	event.Requestor = ID
	event.EventSource = ID

	// construct command and replace
	replacer := strings.NewReplacer(
		"{IP}",
		player.IP,
		"{ID}",
		fmt.Sprintf("%d", player.ID),
		"{DURATION:MINUTES}",
		fmt.Sprintf("%d", int64(cfg.VPNBanDuration/time.Minute)),
		"{DURATION:SECONDS}",
		fmt.Sprintf("%d", int64(cfg.VPNBanDuration/time.Second)),
		"{REASON}",
		reason,
	)
	banCommand := replacer.Replace(cfg.DefaultBanCommand)
	event.Command = banCommand

	if cfg.BroadcastBans {
		// ban on all servers
		publisher.Publish(mqtt.Message{
			Topic:   mqtt.TopicBroadcast,
			Payload: event.Marshal(),
		})
	} else {
		// only ban on the server where the player joined
		publisher.Publish(mqtt.Message{
			Topic:   source,
			Payload: event.Marshal(),
		})
	}
	return nil
}
