package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"internet-shop/repository"
	"log"
)

func consumeOrderMessage(repo repository.ProductRepository, ch *amqp.Channel) {
	message, err := ch.Consume(
		"order_queue",
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

			products, ok := messageData["products"].([]interface{})
			if !ok {
				log.Println("Failed to parse products from message")
				continue
			}

			for _, p := range products {
				productData, ok := p.(map[string]interface{})
				if !ok {
					log.Println("Failed to parse product data")
					continue
				}

				id, ok := productData["id"].(float64)
				if !ok {
					log.Println("Failed to parse product ID")
					continue
				}
				quantity, ok := productData["quantity"].(float64)
				if !ok {
					log.Println("Failed to parse product quantity")
					continue
				}

				err := repo.UpdateProductQuantity(int64(id), int32(quantity))
				if err != nil {
					log.Printf("Error while updating product quantity: %s", err)
				}
			}
		}
	}()
}
