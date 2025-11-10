package pubsub

import (
    "context"
    "encoding/json"
    "log/slog"

    amqp "github.com/rabbitmq/amqp091-go"
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
