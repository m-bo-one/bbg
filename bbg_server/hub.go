package main

import (
	"os"
	"os/signal"
	"sync"

	sarama "gopkg.in/Shopify/sarama.v1"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"

	pb "github.com/DeV1doR/bbg/bbg_server/protobufs"
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *pb.BBGProtocol

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// db client
	DBClient

	sync.RWMutex
}

func NewHub(dbClient DBClient) *Hub {
	return &Hub{
		broadcast:  make(chan *pb.BBGProtocol),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		DBClient:   dbClient,
	}
}

func (h *Hub) sendToPushService(topic string, key string, value string) {
	h.kafkaProducer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}
}

func (h *Hub) tankMutator(key []byte, value []byte) error {
	pbMsg := &pb.Tank{}
	if err := proto.Unmarshal(value, pbMsg); err != nil {
		return err
	}
	var found *Tank
	h.RLock()
	{
		for client, ok := range h.clients {
			if ok && client.tank != nil && client.tank.ID == string(key) {
				log.Infof("Sucker: Tank found #%s \n", client.tank.ID)
				found = client.tank
				break
			}
		}
	}
	h.RUnlock()

	if found != nil {
		found.Update(pbMsg)
		found.WSClient.sendProtoData(pb.BBGProtocol_TTankUpdate, found.ToProtobuf(), true)
	}
	return nil
}

func (h *Hub) listenPushService(topic string, partition int32) error {
	partitionConsumer, err := h.kafkaConsumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer partitionConsumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	log.Infof("Sucker: started...-_-")

	consumed := 0

SuckerLoop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			log.Infof("Sucker: Consumed message key: %s; value: %s\n", msg.Key, msg.Value)
			if msg.Key != nil && msg.Value != nil {
				if err := h.tankMutator(msg.Key, msg.Value); err != nil {
					log.Errorf("Sucker: Tank mutator error: %s\n", err)
				}
			}
			consumed++
		case <-signals:
			os.Exit(1)
			break SuckerLoop
		}
	}

	log.Infof("Sucker: Total consumption: %d\n", consumed)

	return nil
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Infoln("Hub: Client subscribed")
			log.Infof("Hub: Total clients: %+v\n", h.clients)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Infoln("Hub: Client unsubscribed")
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					log.Infoln("Hub: Client closed conn")
				}
			}
		}
	}
}
