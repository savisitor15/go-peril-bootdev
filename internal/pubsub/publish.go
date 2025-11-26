package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	err := publish(
		ch,
		exchange,
		key,
		val,
		"application/json",
		func(val T) ([]byte, error) {
			return json.Marshal(val)
		},
	)
	return err
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	err := publish(
		ch,
		exchange,
		key,
		val,
		"application/gob",
		func(val T) ([]byte, error) {
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(val)
			if err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		})
	return err
}

func publish[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T,
	cont string,
	marshaller func(T) ([]byte, error),
) error {
	data, err := marshaller(val)
	if err != nil {
		return err
	}

	payload := amqp.Publishing{
		ContentType: cont,
		Body:        data,
	}
	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		payload,
	)
}
