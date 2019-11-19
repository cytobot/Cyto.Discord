package main

import (
	"encoding/json"
	"fmt"
	"log"

	nats "github.com/nats-io/nats.go"
)

type NatsClient struct {
	client        *nats.Conn
	subscriptions []*nats.Subscription
}

//WorkerMessage is the expected data to send to a discord worker
type WorkerMessage struct {
	Type      string
	Command   string
	ChannelID string
	GuildID   string
	UserID    string
	SourceID  string
	Payload   interface{}
}

//ListenerQuery is the expected data to send to a discord listener
type ListenerQuery struct {
	Type  string
	Value string
}

//NewNatsClient creates a new NATS client
func NewNatsClient(endpoint string) (*NatsClient, error) {
	client, err := nats.Connect(endpoint)
	if err != nil {
		return nil, err
	}

	//defer client.Drain()

	return &NatsClient{
		client: client,
	}, nil
}

//Publish sends a message to subject subscriptions
func (c *NatsClient) Publish(subject string, msg interface{}) error {
	bytes, err := json.Marshal(msg)

	if err != nil {
		log.Printf("Failed to marshal message: %s", err)
		return err
	}

	err = c.client.Publish(subject, bytes)

	if err != nil {
		log.Printf("Failed to publish message: %s", err)
		return err
	}

	return nil
}

//Subscribe to a nats subject
func (c *NatsClient) Subscribe(subject string) <-chan *nats.Msg {
	channel := make(chan *nats.Msg, 1)

	sub, err := c.client.ChanSubscribe(subject, channel)

	if err != nil {
		panic(fmt.Sprintf("[NATS client error] %s", err))
	}
	defer sub.Drain()

	c.subscriptions = append(c.subscriptions, sub)

	return channel
}

//QueueSubscribe to a nats subject and queue group
func (c *NatsClient) QueueSubscribe(subject string, queue string) <-chan *nats.Msg {
	channel := make(chan *nats.Msg, 1)

	sub, err := c.client.QueueSubscribeSyncWithChan(subject, queue, channel)

	if err != nil {
		panic(fmt.Sprintf("[NATS client error] %s", err))
	}
	defer sub.Drain()

	c.subscriptions = append(c.subscriptions, sub)

	return channel
}

//Shutdown gracefull cleans up subscriptions and the client
func (c *NatsClient) Shutdown() {
	for _, sub := range c.subscriptions {
		sub.Drain()
	}

	c.client.Drain()
}
