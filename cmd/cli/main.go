package main

import (
	"log"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/players"
)

func main() {
	players := players.SomePlayers()
	game, err := shed.New("some-id", "creator-id", players, nil)
	if err != nil {
		log.Fatal("Could not initialise a new game")
	}

	if err := game.Start(); err != nil {
		log.Fatal("Could not start game")
	}
}
