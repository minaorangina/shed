package main

import (
	"log"
	"os"

	engine "github.com/minaorangina/shed/gameengine"
	"github.com/minaorangina/shed/players"
)

func main() {

	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)

	ge, err := engine.New([]*players.Player{player1, player2}, engine.HandleInitialCards)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()
}
