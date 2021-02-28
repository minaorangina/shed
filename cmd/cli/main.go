package main

import (
	"log"

	"github.com/minaorangina/shed/engine"
)

func main() {
	players := engine.SomePlayers()
	ge, err := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    "some-id",
		CreatorID: "creator-id",
		Players:   players,
	})

	if err != nil {
		log.Fatal("could not start game")
	}

	if err := ge.Start(); err != nil {
		log.Fatal("Could not start game")
	}
}
