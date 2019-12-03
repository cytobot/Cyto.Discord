package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	cytonats "github.com/cytobot/messaging/nats"
	pbd "github.com/cytobot/messaging/transport/discord"
	pbm "github.com/cytobot/messaging/transport/manager"
	pbs "github.com/cytobot/messaging/transport/shared"
	pbw "github.com/cytobot/messaging/transport/worker"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/lampjaw/discordgobot"
	"github.com/nats-io/nats.go"
)

type NatsManager struct {
	client        *cytonats.NatsClient
	listenerState *listenerState
	shutdownChan  chan int32
}

var statsStartTime = time.Now()

func NewNatsManager(endpoint string, state *listenerState) (*NatsManager, error) {
	client, err := cytonats.NewNatsClient(endpoint)
	if err != nil {
		return nil, err
	}

	return &NatsManager{
		client:        client,
		listenerState: state,
		shutdownChan:  make(chan int32),
	}, nil
}

// TODO: Keep track of command configurations.
func (m *NatsManager) StartCommandUpdateListener() error {
	subChan, err := m.client.ChanSubscribe("command-update")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-subChan:
				updatedCommandConfigurations := &pbm.UpdatedCommandConfigurations{}
				json.Unmarshal(msg.Data, updatedCommandConfigurations)
			case <-m.shutdownChan:
				return
			}
		}
	}()

	return nil
}

func (m *NatsManager) StartDiscordInformationListener() error {
	subChan, err := m.client.ChanSubscribe(m.listenerState.id)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-subChan:
				req := &pbw.DiscordInformationRequest{}
				json.Unmarshal(msg.Data, req)

				payload := make(map[string]string, 0)

				switch req.Type {
				case "channel":
					channel, err := m.listenerState.bot.Client.DiscordClient.Channel(req.Payload["channelID"])
					if err != nil {
						payload["error"] = fmt.Sprint(err)
						break
					}

					bytes, _ := json.Marshal(channel)
					payload["channel"] = string(bytes)
				case "guild":
					guild, err := m.listenerState.bot.Client.DiscordClient.Guild(req.Payload["guildID"])
					if err != nil {
						payload["error"] = fmt.Sprint(err)
						break
					}

					bytes, _ := json.Marshal(guild)
					payload["guild"] = string(bytes)
				case "user":
					user, err := m.listenerState.bot.Client.DiscordClient.Session.User(req.Payload["userID"])
					if err != nil {
						payload["error"] = fmt.Sprint(err)
						break
					}

					bytes, _ := json.Marshal(user)
					payload["user"] = string(bytes)

					if req.Payload["channelID"] != "" {
						payload["nickname"] = m.listenerState.bot.Client.DiscordClient.NicknameForID(user.ID, user.Username, req.Payload["channelID"])
					}
				}

				m.sendRequestResponse(msg, payload)

			case <-m.shutdownChan:
				return
			}
		}
	}()

	return nil
}

func (m *NatsManager) sendRequestResponse(msg *nats.Msg, payload map[string]string) error {
	data := &pbd.DiscordInformationResponse{
		Timestamp:  mapToProtoTimestamp(time.Now().UTC()),
		InstanceID: m.listenerState.id,
		Payload:    payload,
	}
	return m.client.Publish(msg.Reply, data)
}

func (m *NatsManager) SendWorkerMessage(group string, cmd string, msg discordgobot.Message, parameters map[string]string) {
	guildID, _ := msg.ResolveGuildID()

	content := &pbs.DiscordWorkRequest{
		Type:      group,
		Command:   cmd,
		MessageID: msg.MessageID(),
		ChannelID: msg.Channel(),
		GuildID:   guildID,
		UserID:    msg.UserID(),
		SourceID:  m.listenerState.id,
		Payload:   parameters,
	}

	m.client.Publish("discord_work", content)
}

func (m *NatsManager) StartHealthCheckInterval() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHealthMessage(m)
			case <-m.shutdownChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func sendHealthMessage(m *NatsManager) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	content := &pbd.HealthCheckStatus{
		Timestamp:        mapToProtoTimestamp(time.Now().UTC()),
		InstanceID:       m.listenerState.id,
		ShardID:          int32(m.listenerState.shardID),
		Uptime:           time.Now().Sub(statsStartTime).Nanoseconds(),
		MemAllocated:     int64(stats.Alloc),
		MemSystem:        int64(stats.Sys),
		MemCumulative:    int64(stats.TotalAlloc),
		TaskCount:        int32(runtime.NumGoroutine()),
		ConnectedServers: int32(m.listenerState.bot.Client.ChannelCount()),
		ConnectedUsers:   int32(m.listenerState.bot.Client.UserCount()),
	}

	m.client.Publish("listener_health", content)
}

func (m *NatsManager) Shutdown() {
	m.shutdownChan <- 1
	m.client.Shutdown()
}

func mapToProtoTimestamp(timeValue time.Time) *timestamp.Timestamp {
	protoTimestamp, _ := ptypes.TimestampProto(timeValue)
	return protoTimestamp
}
