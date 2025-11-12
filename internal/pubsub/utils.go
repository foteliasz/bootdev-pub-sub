package pubsub

import (
    "fmt"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
    amqp "github.com/rabbitmq/amqp091-go"
)

func PublishLog(
    channel *amqp.Channel,
    username string,
    message string) error {

    err := PublishGob(
        channel,
        routing.ExchangePerilTopic,
        fmt.Sprintf("%s.%s", routing.GameLogSlug, username),
        message)
    if err != nil {
        return err
    }
    return nil
}
