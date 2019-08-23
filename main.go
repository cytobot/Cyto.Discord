package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	inviteplugin "github.com/cytobot/Cyto.Plugins/invite"
	statsplugin "github.com/cytobot/Cyto.Plugins/stats"
	"github.com/lampjaw/discordgobot"
)

// VERSION is the application version
const VERSION = "0.1.0"

func main() {
	token := os.Getenv("DiscordToken")

	if token == "" {
		fmt.Println("No token provided.")
		return
	}

	ownerUserID := os.Getenv("DiscordOwnerId")
	clientID := os.Getenv("DiscordClientId")

	config := &discordgobot.GobotConf{
		OwnerUserID: ownerUserID,
		ClientID:    clientID,
	}

	bot, err := discordgobot.NewBot(token, config)

	if err != nil {
		log.Printf("Unable to create bot: %s", err)
		return
	}

	bot.RegisterPlugin(inviteplugin.New())
	bot.RegisterPlugin(statsplugin.New(VERSION, true))

	bot.Open()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			bot.Client.Session.Close()
			break out
		}
	}
}
