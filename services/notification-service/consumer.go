package main

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	"internet-shop/services/notification-service/handlers"
	"internet-shop/services/notification-service/senders"
	"log"
)

func consumeNotificationMessage(eSend email.EmailSender, handler handlers.NotificationHandler, ch *amqp.Channel) {
	message, err := ch.Consume(
		"notification_after_order",
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

			orderID, ok := messageData["order_id"].(float64)
			if !ok {
				log.Println("Failed to parse order_id from message")
				continue
			}

			author, ok := messageData["author"].(string)
			if !ok {
				log.Println("Failed to parse author from message")
				continue
			}

			subject, ok := messageData["subject"].(string)
			if !ok {
				log.Println("Failed to parse subject from message")
				continue
			}

			content, ok := messageData["content"].(string)
			if !ok {
				log.Println("Failed to parse content from message")
				continue
			}
			err = handler.SaveNotificaion(context.Background(), int64(userID), author, subject, content)
			err = eSend.SendEmail(int64(userID), int64(orderID), subject, content, handler)
			if err != nil {
				log.Printf("Error while sending email: %s", err)
			} else {
				log.Printf("Notification sent to %s", userID)
			}
		}
	}()
}
