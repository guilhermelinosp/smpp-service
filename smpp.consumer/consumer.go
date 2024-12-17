package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/streadway/amqp"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Request struct {
	To       string `json:"to"`
	Message  string `json:"message"`
	Priority string `json:"priority"`
}

type Message struct {
	Request Request `json:"request"`
	ID      string  `json:"id"`
}

type RabbitConsumer struct {
	connection *amqp.Connection
	queue      string
}

func NewRabbitConsumer(conn *amqp.Connection, queue string) *RabbitConsumer {
	return &RabbitConsumer{
		connection: conn,
		queue:      queue,
	}
}

func (c *RabbitConsumer) Consume() error {
	channel, err := c.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	msgs, err := channel.Consume(
		c.queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for d := range msgs {
		wg.Add(1)
		go func(d amqp.Delivery) {
			defer wg.Done()
			var message Message
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Error decoding message: %v", err)
				return
			}
			c.handleMessage(message.Request, message.ID)
		}(d)
	}

	wg.Wait()

	log.Printf("Consumer started for queue: %s", c.queue)
	return nil
}

func (c *RabbitConsumer) handleMessage(msg Request, id string) {
	if msg.To == "" || msg.Message == "" {
		log.Printf("Invalid message: %+v", msg)
		return
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})

	if client == nil {
		log.Printf("Failed to initialize Twilio client")
		return
	}

	go func() {
		params := &twilioApi.CreateMessageParams{}
		params.SetTo(msg.To)
		params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
		params.SetBody(msg.Message + "\n\nMessage ID: " + id)
		_, err := client.Api.CreateMessage(params)
		if err != nil {
			log.Fatalf("Failed to send SMS: %v", err)
		}

		if err != nil {
			log.Printf("Failed to send SMS. ID: %s, Error: %v", id, err)
			return
		}
		log.Printf("SMS sent successfully. ID: %s", id)
	}()
}
