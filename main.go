package main

import (
	"log"
	"net/http"
	"os"

	engine "github.com/minaorangina/shed/gameengine"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/server"
)

func main() {

	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)

	ge, err := engine.New("", []*players.Player{player1, player2}, engine.HandleInitialCards)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()

	s := server.NewServer(nil)
	http.ListenAndServe(":8000", s)
}
