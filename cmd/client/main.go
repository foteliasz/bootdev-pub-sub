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
		fmt.Println("client is shutting down ")
		if err != nil {
			slog.Error(err.Error())
		}
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	fmt.Println("connected successfully")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		"peril_direct",
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient,
		handlerMove(gameState))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				slog.Error(err.Error())
				continue
			}

		case "move":
			move, err := gameState.CommandMove(input)
			if err != nil {
				slog.Error(err.Error())
				continue
			}

			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				move)
			if err != nil {
				slog.Error(err.Error())
				continue
			}

			fmt.Println("move published successfully")

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()

		default:
			fmt.Println("Unknown command: " + command)

		}
	}
}
