package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	AckTypeAck Acktype = iota
	AckTypeNackRequeue
	AckTypeNackDiscard
)

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
	deadLetter bool) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel %v", err)
	}
	var deadLetterTable amqp.Table
	if deadLetter {
		deadLetterTable = amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		}
	} else {
		deadLetterTable = nil
	}
	queue, err := ch.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable, // check if durable
		queueType != SimpleQueueDurable, // delete if not used
		queueType != SimpleQueueDurable, // exclusive mode?
		false,                           // R-T/ no-wait
		deadLetterTable,                 // dead-letter-queue
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	deadLetter bool,
	handler func(T) Acktype,
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		deadLetter)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	unmarshaller := func(data []byte) (T, error) {
		var payload T
		err := json.Unmarshal(data, &payload)
		return payload, err
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal msg: %v\n", err)
				continue
			}
			ret := handler(target)
			switch ret {
			case AckTypeAck:
				msg.Ack(false)
				log.Printf("message: %v Acknowledged", msg.AppId)
			case AckTypeNackRequeue:
				msg.Nack(false, true)
				log.Printf("message: %v Neg Acknowledge, requeued", msg.AppId)
			case AckTypeNackDiscard:
				msg.Nack(false, false)
				log.Printf("message: %v Neg Acknowledge, discarded", msg.AppId)
			default:
				msg.Nack(false, false)
				log.Printf("message: %v resulted in an unexpected status: %v", msg.AppId, ret)
			}
			msg.Ack(false)
		}
	}()
	return nil
}
