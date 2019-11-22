package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	inviteplugin "github.com/cytobot/Cyto.Plugins/invite"
	statsplugin "github.com/cytobot/Cyto.Plugins/stats"
	"github.com/lampjaw/discordgobot"
	"github.com/lithammer/shortuuid"
)

// VERSION is the application version
const VERSION = "0.1.0"

type listenerState struct {
	id            string
	shardID       int
	bot           *discordgobot.Gobot
	nats          *NatsManager
	managerclient *ManagerClient
}

func main() {
	listener := &listenerState{
		id:            shortuuid.New(),
		shardID:       getShardID(),
		managerclient: getManagerClient(),
	}

	listener.nats = getNatsManager(listener)

	//definitions, err := listener.managerclient.GetCommandDefinitions()

	bot := getDiscordBot(listener)

	bot.RegisterPlugin(inviteplugin.New())
	bot.RegisterPlugin(statsplugin.New(VERSION, true))

	bot.Open()

	listener.bot = bot

	go listener.nats.StartHealthCheckInterval()
	go listener.nats.StartCommandUpdateListener()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			log.Println("Shutting down...")
			bot.Client.Session.Close()
			listener.nats.Shutdown()
			break out
		}
	}
}

func getShardID() int {
	envShardID := os.Getenv("DiscordToken")
	if envShardID != "" {
		shardId, _ := strconv.ParseInt(envShardID, 10, 64)
		return int(shardId)
	}
	return -1
}

func getDiscordBot(state interface{}) *discordgobot.Gobot {
	token := os.Getenv("DiscordToken")

	if token == "" {
		panic("No token provided.")
	}

	ownerUserID := os.Getenv("DiscordOwnerId")
	clientID := os.Getenv("DiscordClientId")

	config := &discordgobot.GobotConf{
		OwnerUserID: ownerUserID,
		ClientID:    clientID,
	}

	bot, err := discordgobot.NewBot(token, config, state)

	if err != nil {
		panic(fmt.Sprintf("Unable to create bot: %s", err))
	}

	return bot
}

func getNatsManager(state *listenerState) *NatsManager {
	natsEndpoint := os.Getenv("NatsEndpoint")

	if natsEndpoint == "" {
		panic("No nats endpoint provided.")
	}

	client, err := NewNatsManager(natsEndpoint, state)
	if err != nil {
		panic(fmt.Sprintf("[NATS error] %s", err))
	}

	log.Println("Connected to NATS")

	return client
}

func getManagerClient() *ManagerClient {
	managerEndpoint := os.Getenv("ManagerEndpoint")

	if managerEndpoint == "" {
		panic("No manager endpoint provided.")
	}

	client, err := NewManagerClient(managerEndpoint)
	if err != nil {
		panic(fmt.Sprintf("[Manager client error] %s", err))
	}

	log.Println("Connected to manager client")

	return client
}
