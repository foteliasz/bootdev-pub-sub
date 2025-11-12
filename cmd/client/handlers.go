package main

import (
	"fmt"
	"log/slog"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Println("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, channel *amqp.Channel) func(outcome gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Println("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				fmt.Sprintf(
					"%s.%s",
					routing.WarRecognitionsPrefix,
					gs.GetUsername()),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				slog.Error(err.Error())
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(channel *amqp.Channel, gs *gamelogic.GameState) func(war gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(row gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Println("> ")

		outcome, winner, loser := gs.HandleWar(row)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeYouWon:
		case gamelogic.WarOutcomeOpponentWon:
			message := fmt.Sprintf(
				"%s won a war against %s",
				winner,
				loser)

			slog.Info(message)
			err := pubsub.PublishLog(channel, row.Attacker.Username, message)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			message := fmt.Sprintf(
				"A war between %s and %s resulted in a draw",
				winner,
				loser)
			slog.Info(message)
			err := pubsub.PublishLog(channel, row.Attacker.Username, message)
			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		slog.Error(fmt.Sprintf("unsupported war outcome: %d", outcome))
		return pubsub.NackDiscard
	}
}
