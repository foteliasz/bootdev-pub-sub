package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		fmt.Println("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(outcome gamelogic.ArmyMove) {
	return func(outcome gamelogic.ArmyMove) {
		fmt.Println("> ")
		gs.HandleMove(outcome)
	}
}
