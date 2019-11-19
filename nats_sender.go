package main

import "github.com/lampjaw/discordgobot"

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
