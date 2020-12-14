package main

import (
	"log"

	"github.com/minaorangina/shed"
)

func main() {
	players := shed.SomePlayers()
	game, err := shed.NewGameEngine(shed.GameEngineOpts{
		GameID:    "some-id",
		CreatorID: "creator-id",
		Players:   players,
	})

	if err != nil {
		log.Fatal("could not start game")
	}

	if err := game.Start(); err != nil {
		log.Fatal("Could not start game")
	}
}
