package main

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"internet-shop/repository"
	"log"
)

func consumeUserMessage(repo repository.CartRepository, ch *amqp.Channel) {
	message, err := ch.Consume(
		"user_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error while consuming message: %s", err)
	}

	go func() {
		for msg := range message {
			var messageData map[string]interface{}
			err := json.Unmarshal(msg.Body, &messageData)
			if err != nil {
				log.Printf("Error while unmarshalling message: %s", err)
				continue
			}

			userID, ok := messageData["user_id"].(float64)
			if !ok {
				log.Println("Failed to parse user_id from message")
				continue
			}

			err = repo.CreateCart(context.Background(), int64(userID))
			if err != nil {
				log.Printf("Error while creating new cart: %s", err)
			}
		}
	}()
}
