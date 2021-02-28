package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/minaorangina/shed/server"
	"github.com/minaorangina/shed/store"
)

func main() {
	s := server.NewServer(store.NewInMemoryGameStore())
	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", handlers.CombinedLoggingHandler(os.Stdout, s)))
}
