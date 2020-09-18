package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/minaorangina/shed"
	"github.com/minaorangina/shed/players"
	uuid "github.com/satori/go.uuid"
)

type NewGameReq struct {
	Name string `json:"name"`
}

type NewGameRes struct {
	GameID   string `json:"game_id"`
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

type JoinGameReq struct {
	GameID string `json:"game_id"`
	Name   string `json:"name"`
}

type JoinGameRes struct {
	PlayerID string `json:"player_id"`
}

// GameServer is a game server
type GameServer struct {
	store shed.GameStore
	http.Handler
}

func NewID() string {
	return uuid.NewV4().String()
}

// NewServer creates a new GameServer
func NewServer(store shed.GameStore) *GameServer {
	s := new(GameServer)

	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Root endpoint")
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
		}
		w.WriteHeader(http.StatusOK)
	}))
	router.Handle("/new", http.HandlerFunc(s.HandleNewGame))
	router.Handle("/game/", http.HandlerFunc(s.HandleFindGame))
	router.Handle("/join", http.HandlerFunc(s.HandleJoinGame))

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

	if err != nil {
		writeParseError(err, w, r)
		return
	}

	// generate game ID
	gameID := NewID()
	playerID := NewID()
	creator := players.NewPlayer(playerID, data.Name, &bytes.Buffer{}, ioutil.Discard)

	err = g.store.AddPendingGame(gameID, creator)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := NewGameRes{gameID, playerID, data.Name}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (g *GameServer) HandleFindGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	gameID := strings.Replace(r.URL.String(), "/game/", "", 1)
	if gameID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var status string

	_, ok := g.store.FindGame(gameID)
	if ok {
		status = "active"
	} else {
		_, ok := g.store.FindPendingPlayers(gameID)
		if ok {
			status = "pending"
		}
	}

	if status == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{"status": "` + status + `", "game_id": "` + gameID + `"}`))
}

func (g *GameServer) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
	var data JoinGameReq
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close() // why?

	if err != nil {
		writeParseError(err, w, r)
		return
	}

	if data.GameID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing game ID"))
		return
	}

	if data.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing player name"))
		return
	}

	// identify game
	_, ok := g.store.FindPendingPlayers(data.GameID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Game matching id '%s' not found", data.GameID)))
		return
	}

	// init websocket?

	// make player
	joiningPlayerID := NewID()
	joiningPlayer := players.NewPlayer(joiningPlayerID, data.Name, &bytes.Buffer{}, ioutil.Discard)

	err = g.store.AddToPendingPlayers(data.GameID, joiningPlayer)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to add new player to game: %v", err)))
		return
	}

	payload := JoinGameRes{joiningPlayerID}
	bytes, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func writeParseError(err error, w http.ResponseWriter, r *http.Request) {
	if err == io.EOF {
		log.Println(err.Error())
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing body"))
		return
	}
	if err != nil {
		log.Println(err.Error())
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
