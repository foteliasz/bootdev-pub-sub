package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"log/slog"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type AckType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T) error {

	data, err := json.Marshal(val)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func PublishGob[T any](
	ch *amqp.Channel,
	exchange,
	key string,
	val T) error {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		log.Fatal(err)
	}

	err := ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buf.Bytes(),
		})
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange string,
	queueName string,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		slog.Error(err.Error())
		return nil, amqp.Queue{}, err
	}

	table := amqp.Table{}
	table["x-dead-letter-exchange"] = routing.ExchangePerilFanout
	queue, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		table)
	if err != nil {
		slog.Error(err.Error())
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange string,
	queueName string,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	amqpChannel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	channel, err := amqpChannel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	go func() {
		for delivery := range channel {
			var object T
			err := json.Unmarshal(delivery.Body, &object)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			ackType := handler(object)
			switch ackType {
			case Ack:
				err = delivery.Ack(false)
				slog.Info("message acknowledged")
				if err != nil {
					slog.Error(err.Error())
					return
				}

			case NackRequeue:
				err = delivery.Nack(false, true)
				slog.Info("message not acknowledged, queuing again")
				if err != nil {
					slog.Error(err.Error())
					return
				}

			case NackDiscard:
				err = delivery.Nack(false, false)
				slog.Info("message not acknowledged, discarding")
				if err != nil {
					slog.Error(err.Error())
					return
				}
			}
		}
	}()

	return nil
}
