package main

import (
	"encoding/json"
)

func (listener *listenerState) setupDiscordInfoSubscription() {
	listenerInfo := listener.nats.Subscribe(listener.id)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-listenerInfo:

				var data ListenerQuery
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					go listener.nats.Publish(msg.Reply, err)
					return
				}

				var results interface{}

				switch data.Type {
				case "guild":
					guild, err := listener.bot.Client.Guild(data.Value)
					if err != nil {
						results = err
					} else {
						results = guild
					}
				case "channel":
					channel, err := listener.bot.Client.Channel(data.Value)
					if err != nil {
						results = err
					} else {
						results = channel
					}
				}

				go listener.nats.Publish(msg.Reply, results)
			case <-quit:
				return
			}
		}
	}()
}
