package pubsub

import (
    "context"
    "encoding/json"
    "log/slog"

    amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
    Durable SimpleQueueType = iota
    Transient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
    data, err := json.Marshal(val)
    if err != nil {
        slog.Error(err.Error())
        return err
    }

    err = ch.PublishWithContext(context.Background(), exchange, key,
        false, false, amqp.Publishing{
            ContentType: "application/json",
            Body:        data,
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

    queue, err := channel.QueueDeclare(
        queueName,
        queueType == Durable,
        queueType == Transient,
        queueType == Transient,
        false,
        nil)
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
    handler func(T),
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

            handler(object)
            err = delivery.Ack(false)
            if err != nil {
                slog.Error(err.Error())
                return
            }
        }
    }()

    return nil
}
