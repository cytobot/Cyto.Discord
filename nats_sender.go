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

func (l *listenerState) SendHealthMessage() {
	content := &HealthMessage{
		ID:        l.id,
		Timestamp: time.Now().UTC(),
	}

	l.nats.Publish("listener_health", content)
}
