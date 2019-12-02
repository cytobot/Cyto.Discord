package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/lampjaw/discordgobot"
	"github.com/lithammer/shortuuid"
)

type listenerState struct {
	id             string
	shardID        int
	bot            *discordgobot.Gobot
	nats           *NatsManager
	managerclient  *ManagerClient
	commandMonitor *CommandMonitor
}

func main() {
	listener := &listenerState{
		id:            shortuuid.New(),
		shardID:       getShardID(),
		managerclient: getManagerClient(),
	}

	listener.bot = getDiscordBot(listener)
	listener.nats = getNatsManager(listener)
	listener.commandMonitor = getCommandMonitor(listener.managerclient, listener.bot, listener.nats)

	go listener.nats.StartHealthCheckInterval()
	go listener.nats.StartCommandUpdateListener()

	listener.bot.Open()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			log.Println("Shutting down...")
			listener.bot.Client.Session.Close()
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
		OwnerUserID:   ownerUserID,
		ClientID:      clientID,
		CommandPrefix: "cy?",
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

func getCommandMonitor(client *ManagerClient, bot *discordgobot.Gobot, natsManager *NatsManager) *CommandMonitor {
	monitor, err := NewCommandMonitor(client, bot, natsManager)
	if err != nil {
		panic(fmt.Sprintf("[Command Monitor error] %s", err))
	}

	log.Printf("Retrieved %d command definitions", len(monitor.commandDefinitions))

	return monitor
}
