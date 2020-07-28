package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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
	http.Handler
}

// NewServer creates a new GameServer
func NewServer() *GameServer {
	s := new(GameServer)

	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	router.Handle("/new", http.HandlerFunc(HandleNewGame))

	s.Handler = router

	return s
}

// HandleNewGame handles a request to create a new game
func HandleNewGame(w http.ResponseWriter, r *http.Request) {
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
