package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"

	"github.com/joeshaw/envdecode"
	"github.com/minaorangina/shed/server"
	"github.com/minaorangina/shed/store"
)

type EnvVars struct {
	Port int `env:"PORT,default=8000"`
}

var Env EnvVars

func init() {
	if err := envdecode.Decode(&Env); err != nil {
		log.Fatalf("config: %s", err.Error())
	}
}

func main() {
	s := server.NewServer(store.NewInMemoryGameStore())

	go func() {
		port := fmt.Sprintf(":%d", Env.Port)
		log.Printf("Listening on port %d...", Env.Port)

		if err := http.ListenAndServe(
			port,
			handlers.CombinedLoggingHandler(os.Stdout, s),
		); err != http.ErrServerClosed {
			log.Fatalf("something went wrong: %v", err)
		}
	}()

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	requestGracePeriod := time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), requestGracePeriod)
	defer cancel()

	select {
	case sig := <-signalCh:
		fmt.Printf("Got %v signal, aborting...", sig)

		err := s.Shutdown(ctx)
		if err != nil {
			log.Fatalf("could not gracefully shut down server: %v", err)
		}
	}
}
