package main

import (
	"time"

	"github.com/lampjaw/discordgobot"
)

func (l *listenerState) SendWorkerMessage(group string, cmd string, msg discordgobot.Message, parameters map[string]string) {
	guildID, _ := msg.ResolveGuildID()

	content := &WorkerMessage{
		Type:      group,
		Command:   cmd,
		ChannelID: msg.Channel(),
		GuildID:   guildID,
		UserID:    msg.UserID(),
		SourceID:  l.id,
		Payload:   parameters,
	}

	l.nats.Publish("discord_work", content)
}

func (l *listenerState) setupHealthCheckInterval() {
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHealthMessage(l)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func sendHealthMessage(l *listenerState) {
	content := &HealthMessage{
		ID:        l.id,
		Timestamp: time.Now().UTC(),
	}

	l.nats.Publish("listener_health", content)
}
