package main

import (
	"log"
	"net/http"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/server"
)

func main() {

	player1 := shed.NewWSPlayer(shed.NewID(), "Harry", nil)
	player2 := shed.NewWSPlayer(shed.NewID(), "Sally", nil)

	ps := shed.NewPlayers(player1, player2)

	ge, err := shed.New("", "", ps, shed.HandleInitialCards)
	if err != nil {
		log.Fatal(err.Error())
	}
	ge.Start()

	s := server.NewServer(shed.NewInMemoryGameStore())
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", s))
	// log.Fatal(s.ListenAndServe(":8000"))
}
