package main

import (
	"log"
	"net/http"
	"os"

	engine "github.com/minaorangina/shed"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/server"
)

func main() {

	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)

	ps := players.NewPlayers(player1, player2)

	ge, err := engine.New("", ps, engine.HandleInitialCards)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()

	s := server.NewServer(nil)
	log.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8000", s))
}
