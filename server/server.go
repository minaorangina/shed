package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed"
	uuid "github.com/satori/go.uuid"
)

var (
	homepage            = "./build/index.html"
	waitingRoomTemplate = "./static/waiting-room.tmpl"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NewGameReq struct {
	Name string `json:"name"`
}

type PendingGameRes struct {
	GameID   string   `json:"game_id"`
	PlayerID string   `json:"player_id"`
	Name     string   `json:"name"`
	Admin    bool     `json:"is_admin"`
	Players  []string `json:"players"`
}

type JoinGameReq struct {
	GameID string `json:"game_id"`
	Name   string `json:"name"`
}
type GetGameRes struct {
	Status string `json:"status"`
	GameID string `json:"game_id"`
}

// GameServer is a game server
type GameServer struct {
	store shed.GameStore
	http.Server
}

func NewID() string {
	return uuid.NewV4().String()
}

func NewGameID() string {
	letters := []byte("ABCDEFGHIJKLMNOPQURSTUVWXYZ")
	var code = []byte{}

	for i := 0; i < 6; i++ {
		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(25)
		code = append(code, letters[idx])
	}

	return string(code)
}

func unknownGameIDMsg(unknownID string) string {
	return fmt.Sprintf("unknown game ID '%s'", unknownID)
}

func enableCors(handler http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		handler.ServeHTTP(w, r)
	}
}

func servePage(w http.ResponseWriter, path string) {
	log.Print(path)
	tmpl, err := template.ParseFiles(path)

	if err != nil {
		http.Error(w, fmt.Sprintf("problem loading template %s", err.Error()), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// NewServer creates a new GameServer
func NewServer(store shed.GameStore) *GameServer {
	s := new(GameServer)

	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Root endpoint")
		if r.URL.Path != "/" {
			log.Println(r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		servePage(w, homepage)
	}))

	fileServer := http.FileServer(http.Dir("./build"))
	router.Handle("/static/", fileServer)
	router.Handle("/build/", http.StripPrefix("/build/", fileServer))
	router.Handle("/new", http.HandlerFunc(enableCors(s.HandleNewGame)))
	router.Handle("/game/", http.HandlerFunc(s.HandleFindGame))
	router.Handle("/join", http.HandlerFunc(enableCors(s.HandleJoinGame)))
	router.Handle("/waiting-room", http.HandlerFunc(s.HandleWaitingRoom))
	router.Handle("/ws", http.HandlerFunc(enableCors(s.HandleWS)))

	s.store = store

	s.Handler = router

	return s
}

// ServeHTTP serves http
func (g *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Handler.ServeHTTP(w, r)
}

// HandleNewGame handles a request to create a new game
func (g *GameServer) HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var data NewGameReq
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close()
	if err != nil {
		writeParseError(err, w, r)
		return
	}

	// generate game ID
	gameID := NewGameID()
	playerID := NewID()
	game, err := shed.NewGameEngine(shed.GameEngineOpts{
		GameID:    gameID,
		CreatorID: playerID,
		Game:      shed.NewShed(shed.ShedOpts{}),
	})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if game == nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get hub running
	go game.Listen()

	err = g.store.AddInactiveGame(game)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = g.store.AddPendingPlayer(gameID, playerID, data.Name)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := PendingGameRes{
		GameID:   gameID,
		PlayerID: playerID,
		Name:     data.Name,
		Admin:    true,
	}

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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing game ID"))
		return
	}

	gameExists := findGame(g.store)

	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(unknownGameIDMsg(gameID)))
		return
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(responseBytes)
}

func (g *GameServer) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
	var data JoinGameReq
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close()

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

	// This step is repeated in AddPendingPlayer. One of these will have to go eventually.
	game := g.store.FindInactiveGame(data.GameID)
	if game == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(unknownGameIDMsg(data.GameID)))
		return
	}

	playerID := NewID()

	err = g.store.AddPendingPlayer(data.GameID, playerID, data.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	playerNames := []string{}
	ps := game.Players()
	for _, p := range ps {
		playerNames = append(playerNames, p.Name())
	}

	payload := PendingGameRes{
		PlayerID: playerID,
		GameID:   data.GameID,
		Name:     data.Name,
		Players:  playerNames,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (g *GameServer) HandleWaitingRoom(w http.ResponseWriter, r *http.Request) {
	// check if this person should get the file
	query := r.URL.Query()
	vals, ok := query["game_id"]
	if !ok || len(vals) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing game ID"))
		return
	}
	gameID := vals[0]

	vals, ok = query["player_id"]
	if !ok || len(vals) != 1 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("missing user ID"))
		return
	}

	playerID := vals[0]
	_ = playerID

	game := g.store.FindInactiveGame(gameID)
	if game == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(unknownGameIDMsg(gameID)))
		return
	}

	tmpl, err := template.ParseFiles(waitingRoomTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong"))
		return
	}

	data := struct {
		GameID  string
		IsAdmin bool
	}{
		GameID:  gameID,
		IsAdmin: playerID == game.CreatorID(),
	}

	tmpl.Execute(w, data)
}

func (g *GameServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vals, ok := query["game_id"]
	log.Println(vals, ok)
	if !ok || len(vals) != 1 {
		log.Println("missing game ID")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing game ID"))
		return
	}
	gameID := vals[0]

	vals, ok = query["player_id"]
	if !ok || len(vals) != 1 {
		log.Println("missing player ID")
		w.Write([]byte("missing player ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	playerID := vals[0]
	game := g.store.FindInactiveGame(gameID)
	if game == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(unknownGameIDMsg(gameID)))
		return
	}

	pendingPlayer := g.store.FindPendingPlayer(gameID, playerID)
	if pendingPlayer == nil {
		log.Println("missing player ID")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown player ID"))
		return
	}

	rawConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("could not upgrade to websocket: %v", err)))
		return
	}

	// create player
	player := shed.NewWSPlayer(playerID, pendingPlayer.Name, rawConn, make(chan []byte), game)
	// reference to hub etc
	err = game.AddPlayer(player)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("could not add player to game: %v", err)))
		return
	}
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

func findGame(s shed.GameStore) bool {
	game := g.store.FindGame(gameID)
	if game != nil {
		return true
	}
	return false
}
