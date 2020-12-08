package main

import (
	"log"

	"github.com/minaorangina/shed"
)

func main() {
	players := shed.SomePlayers()
	game := shed.NewGameEngine(shed.GameEngineOpts{
		GameID:    "some-id",
		CreatorID: "creator-id",
		Players:   players,
	})

	if err := game.Start(); err != nil {
		log.Fatal("Could not start game")
	}
}
