package main

import (
	"log"
	"net/http"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/server"
)

func main() {
	s := server.NewServer(shed.NewInMemoryGameStore())
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", s))
	// log.Fatal(s.ListenAndServe(":8000"))
}
