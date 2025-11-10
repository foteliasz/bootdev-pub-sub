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
    queueType SimpleQueueType, // an enum to represent "Durable" or "Transient"
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
