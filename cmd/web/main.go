package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/server"
)

func main() {
	s := server.NewServer(shed.NewInMemoryGameStore())
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", handlers.CombinedLoggingHandler(os.Stdout, s)))
}
