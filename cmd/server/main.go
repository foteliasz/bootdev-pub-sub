package main

import (
    "fmt"
    "log/slog"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
        fmt.Println("server is shutting down")
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

    _, _, err = pubsub.DeclareAndBind(
        connection,
        "peril_topic",
        "game_logs",
        "game_logs.*",
        pubsub.Durable)
    if err != nil {
        slog.Error(err.Error())
        return
    }

    gamelogic.PrintServerHelp()

    for {
        input := gamelogic.GetInput()
        if len(input) == 0 {
            continue
        }

        command := input[0]
        switch command {
        case "pause":
            fmt.Println("sending a pause message")
            err = pubsub.PublishJSON(
                channel,
                routing.ExchangePerilDirect,
                routing.PauseKey,
                routing.PlayingState{
                    IsPaused: true,
                })
            if err != nil {
                slog.Error(err.Error())
                return
            }

        case "resume":
            fmt.Println("sending a resume message")
            err = pubsub.PublishJSON(
                channel,
                routing.ExchangePerilDirect,
                routing.PauseKey,
                routing.PlayingState{
                    IsPaused: false,
                })
            if err != nil {
                slog.Error(err.Error())
                return
            }

        case "quit":
            fmt.Println("exiting...")
            break

        default:
            fmt.Println("unknown command")
        }
    }
}
