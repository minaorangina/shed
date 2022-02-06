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
	"github.com/minaorangina/shed/engine"
	"github.com/minaorangina/shed/game"
	"github.com/minaorangina/shed/protocol"
	"github.com/minaorangina/shed/store"
	str "github.com/minaorangina/shed/store"
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
	GameID     string            `json:"gameID"`
	PlayerID   string            `json:"playerID"`
	Name       string            `json:"name"`
	PlayerInfo protocol.Player   `json:"playerInfo"`
	Admin      bool              `json:"isAdmin"`
	Players    []protocol.Player `json:"players,omitempty"`
}

type JoinGameReq struct {
	GameID string `json:"gameID"`
	Name   string `json:"name"`
}
type GetGameRes struct {
	State  string `json:"state"`
	GameID string `json:"gameID"`
}

// GameServer is a game server
type GameServer struct {
	store str.GameStore
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
	tmpl, err := template.ParseFiles(path)

	if err != nil {
		http.Error(w, fmt.Sprintf("problem loading template %s", err.Error()), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

// NewServer creates a new GameServer
func NewServer(str store.GameStore) *GameServer {
	s := new(GameServer)

	router := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./build"))
	router.Handle("/", fileServer)
	router.Handle("/static/", fileServer)
	router.Handle("/build/", http.StripPrefix("/build/", fileServer))
	router.Handle("/new", http.HandlerFunc(enableCors(s.HandleNewGame)))
	router.Handle("/game/", http.HandlerFunc(s.HandleFindGame))
	router.Handle("/join", http.HandlerFunc(enableCors(s.HandleJoinGame)))
	router.Handle("/waiting-room", http.HandlerFunc(s.HandleWaitingRoom))
	router.Handle("/ws", http.HandlerFunc(enableCors(s.HandleWS)))

	s.store = str

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
		writePathNotFoundError(w, fmt.Sprintf("%s %s not found", r.Method, r.URL.Path))
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
	game, err := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    gameID,
		CreatorID: playerID,
		Game:      game.NewShed(game.ShedOpts{}),
	})
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if game == nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = g.store.AddInactiveGame(game)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = g.store.AddPendingPlayer(gameID, playerID, data.Name)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := PendingGameRes{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerInfo: protocol.Player{PlayerID: playerID, Name: data.Name},
		Name:       data.Name,
		Players:    []protocol.Player{{PlayerID: playerID, Name: data.Name}},
		Admin:      true,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		writeMarshalError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (g *GameServer) HandleFindGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writePathNotFoundError(w, fmt.Sprintf("path %s %s not found", r.Method, r.URL.Path))
		return
	}

	gameID := strings.Replace(r.URL.String(), "/game/", "", 1)
	if gameID == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	engine := g.store.FindGame(gameID)

	if engine == nil {
		http.Error(w, unknownGameIDMsg(gameID), http.StatusNotFound)
		return
	}

	game := engine.Game()
	if game == nil {
		http.Error(w, "GameEngine had nil game", http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(game)
	if err != nil {
		writeMarshalError(w, err)
		return
	}

	response := GetGameRes{
		State:  string(bytes),
		GameID: engine.ID(),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		writeMarshalError(w, err)
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
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Missing player name", http.StatusBadRequest)
		return
	}

	// This step is repeated in AddPendingPlayer. One of these will have to go eventually.
	game := g.store.FindInactiveGame(data.GameID)
	if game == nil {
		http.Error(w, unknownGameIDMsg(data.GameID), http.StatusBadRequest)
		return
	}

	playerID := NewID()

	err = g.store.AddPendingPlayer(data.GameID, playerID, data.Name)
	if err != nil {
		log.Println("error adding player to game", err)
		http.Error(w, "Could not add player to game", http.StatusInternalServerError)
		return
	}

	playerInfos := []protocol.Player{}
	ps := game.Players()
	for _, p := range ps {
		playerInfos = append(playerInfos, protocol.Player{
			PlayerID: p.ID(),
			Name:     p.Name(),
		})
	}

	payload := PendingGameRes{
		PlayerID:   playerID,
		GameID:     data.GameID,
		PlayerInfo: protocol.Player{PlayerID: playerID, Name: data.Name},
		Name:       data.Name,
		Players:    playerInfos,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		writeMarshalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (g *GameServer) HandleWaitingRoom(w http.ResponseWriter, r *http.Request) {
	// check if this person should get the file
	query := r.URL.Query()
	vals, ok := query["gameID"]
	if !ok || len(vals) != 1 {
		http.Error(w, "missing game ID", http.StatusBadRequest)
		return
	}
	gameID := vals[0]

	vals, ok = query["playerID"]
	if !ok || len(vals) != 1 {
		http.Error(w, "missing player ID", http.StatusBadRequest)
		return
	}

	playerID := vals[0]
	_ = playerID

	game := g.store.FindInactiveGame(gameID)
	if game == nil {
		http.Error(w, unknownGameIDMsg(gameID), http.StatusUnauthorized)
		return
	}

	tmpl, err := template.ParseFiles(waitingRoomTemplate)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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
	vals, ok := query["gameID"]

	if !ok || len(vals) != 1 {
		log.Println("missing game ID")
		http.Error(w, "missing game ID", http.StatusBadRequest)
		return
	}
	gameID := vals[0]

	vals, ok = query["playerID"]
	if !ok || len(vals) != 1 {
		http.Error(w, "missing player ID", http.StatusInternalServerError)
		return
	}

	playerID := vals[0]
	game := g.store.FindInactiveGame(gameID)
	if game == nil {
		http.Error(w, unknownGameIDMsg(gameID), http.StatusBadRequest)
		return
	}

	pendingPlayer := g.store.FindPendingPlayer(gameID, playerID)
	if pendingPlayer == nil {
		log.Println("unknown player ID")
		http.Error(w, "unknown player ID", http.StatusBadRequest)
		return
	}

	rawConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("could not upgrade to websocket: %v", err), http.StatusInternalServerError)
		return
	}

	// create player
	player := engine.NewWSPlayer(playerID, pendingPlayer.Name, rawConn, make(chan []byte), game)
	// reference to hub etc
	err = game.AddPlayer(player)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("could not add player to game: %v", err), http.StatusInternalServerError)
		return
	}
}

func writeParseError(err error, w http.ResponseWriter, r *http.Request) {
	log.Println(err.Error())
	if err == io.EOF {
		http.Error(w, "Missing body", http.StatusBadRequest)
		return
	}
	http.Error(w, "Could not parse data", http.StatusInternalServerError)
}

func writePathNotFoundError(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusNotFound)
}

func writeMarshalError(w http.ResponseWriter, err error) {
	log.Println("error marshalling json", err)
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}
