package main

import (
    "log/slog"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(log routing.GameLog) pubsub.AckType {
    return func(log routing.GameLog) pubsub.AckType {
        defer slog.Info("received message: " + log.Message)

        err := gamelogic.WriteLog(log)
        if err != nil {
            slog.Error(err.Error())
            return pubsub.NackRequeue
        }
        return pubsub.Ack
    }
}
