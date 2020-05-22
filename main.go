package main

import (
	"log"
	"os"

	engine "github.com/minaorangina/shed/gameengine"
)

func main() {

	player1 := engine.NewPlayer(engine.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := engine.NewPlayer(engine.NewID(), "Sally", os.Stdin, os.Stdout)
	players := engine.NewAllPlayers(player1, player2)

	ge, err := engine.New(players)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()
}
