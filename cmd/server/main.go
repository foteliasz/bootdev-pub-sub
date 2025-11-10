package main

import (
    "fmt"
    "log/slog"
    "os"
    "os/signal"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
    amqp "github.com/rabbitmq/amqp091-go"
)

var connStr = "amqp://guest:guest@localhost:5672/"

func main() {
    connection, err := amqp.Dial(connStr)
    if err != nil {
        slog.Error(err.Error())
        return
    }
    defer func(connection *amqp.Connection) {
        err := connection.Close()
        fmt.Println("program is shutting down ")
        if err != nil {
            slog.Error(err.Error())
        }
    }(connection)

    fmt.Println("connected successfully")

    channel, err := connection.Channel()
    if err != nil {
        slog.Error(err.Error())
        return
    }

    err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect,
        routing.PauseKey, routing.PlayingState{
            IsPaused: true,
        })

    // wait for ctrl+c
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt)
    <-signalChan
}
