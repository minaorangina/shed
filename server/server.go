package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type NewGameReq struct {
	Name string `json:"name"`
}

type NewGameRes struct {
	GameID   string `json:"game_id`
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

// GameServer is a game server
type GameServer struct {
	store GameStore
	http.Handler
}

// NewServer creates a new GameServer
func NewServer(store GameStore) *GameServer {
	s := new(GameServer)

	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Root endpoint")
		w.WriteHeader(http.StatusOK)
	}))
	router.Handle("/new", http.HandlerFunc(s.HandleNewGame))
	router.Handle("/game/", http.HandlerFunc(s.HandleGetGame))

	s.store = store

	s.Handler = router

	return s
}

// HandleNewGame handles a request to create a new game
func (g *GameServer) HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data NewGameReq
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close() // why?

	if err == io.EOF {
		log.Println(err.Error())
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing player name"))
		return
	}
	if err != nil {
		log.Println(err.Error())
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := NewGameRes{"gameid", "playerid", data.Name}
	bytes, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (g *GameServer) HandleGetGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	gameID := strings.Replace(r.URL.String(), "/game/", "", 1)
	if gameID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, ok := g.store.GetGame(gameID); ok {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"game_id": "` + gameID + `"}`))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
