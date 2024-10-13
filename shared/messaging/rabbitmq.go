package messaging

import "github.com/streadway/amqp"

func NewRabbitMQConnections(rabbitURL string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewRabbitMQChannels(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func DeclareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return amqp.Queue{}, err
	}
	return queue, nil
}
