package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/gorilla/websocket"
)

var topic = "web-chat"
var brokers = []string{"127.0.0.1:9092"}
var clients = make(map[*websocket.Conn]struct{}) // connected clients

// Message Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var producer sarama.SyncProducer
var consumer sarama.Consumer

func main() {
	// 生产者配置
	pConfig := sarama.NewConfig()
	pConfig.Producer.Return.Successes = true
	var err error
	if producer, err = sarama.NewSyncProducer(brokers, pConfig); err != nil {
		log.Fatalf("sarama.NewSyncProducer failed: %v", err)
	}
	defer func() {
		if err = producer.Close(); err != nil {
			log.Fatalf("producer.Close failed: %v", err)
		}
	}()

	// 消费者配置
	cConfig := sarama.NewConfig()
	if consumer, err = sarama.NewConsumer(brokers, cConfig); err != nil {
		log.Fatalf("sarama.NewConsumer failed: %v", err)
	}
	defer func() {
		if err = consumer.Close(); err != nil {
			log.Fatalf("consumer.Close failed: %v", err)
		}
	}()

	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := &websocket.Upgrader{}
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("upgrader.Upgrade failed: %v", err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = struct{}{}

	// 循环生产消息
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("ws.ReadJSON failed: %v", err)
			delete(clients, ws)
			break
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("json.Marshal failed: %v", err)
		}
		kfkMessage := sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(data),
		}
		if _, _, err = producer.SendMessage(&kfkMessage); err != nil {
			log.Fatalf("producer.SendMessage failed: %v", err)
		}
	}
}

func handleMessages() {
	// kafka默认只创建一个分区
	// 创建分区0的消费者
	partConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("consumer.ConsumePartition failed: %v", err)
	}

	// 循环消费消息
	for kfkMessage := range partConsumer.Messages() {
		var msg Message
		if err := json.Unmarshal(kfkMessage.Value, &msg); err != nil {
			log.Fatalf("json.Unmarshal failed: %v", err)
		}
		// Send it out to every client that is currently connected
		for client := range clients {
			err = client.WriteJSON(msg)
			if err != nil {
				log.Fatalf("client.WriteJSON error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
