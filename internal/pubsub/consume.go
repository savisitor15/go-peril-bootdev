package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel %v", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable, // check if durable
		queueType != SimpleQueueDurable, // delete if not used
		queueType != SimpleQueueDurable, // exclusive mode?
		false,                           // R-T/ no-wait
		nil,                             // no args
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue %v", err)
	}

	err = ch.QueueBind(
		queue.Name, // queue name
		key,        // routing key
		exchange,   // exchange
		false,      // no-wait / Real-time
		nil,        // args - none
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}
	return ch, queue, nil
}
