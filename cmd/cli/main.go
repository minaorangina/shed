package main

import (
	"log"

	"github.com/minaorangina/shed"
)

func main() {
	players := shed.SomePlayers()
	game, err := shed.NewGameEngine("some-id", "creator-id", players, nil, nil, nil)
	if err != nil {
		log.Fatal("Could not initialise a new game")
	}

	if err := game.Start(); err != nil {
		log.Fatal("Could not start game")
	}
}
