package main

import (
	"log"
	"net/http"
	"os"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/server"
)

func main() {

	player1 := players.NewWSPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewWSPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)

	ps := players.NewPlayers(player1, player2)

	ge, err := shed.New("", ps, shed.HandleInitialCards)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()

	s := server.NewServer(shed.NewInMemoryGameStore(nil, nil))
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", s))
}
