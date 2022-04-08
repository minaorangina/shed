package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed/engine"
	"github.com/minaorangina/shed/game"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
	"github.com/minaorangina/shed/store"
)

type stubStore struct {
	ActiveGames    map[string][]string
	PendingPlayers map[string][]string
	InactiveGames  map[string][]string
}

func (s *stubStore) FindActiveGame(ID string) engine.GameEngine {
	// allows any game
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{Game: game.ExistingShed(game.ShedOpts{})})
	return game
}
func (s *stubStore) FindInactiveGame(ID string) engine.GameEngine {
	// allows any pending game
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{Game: game.ExistingShed(game.ShedOpts{})})
	return game
}

// find pending player

func (s *stubStore) AddInactiveGame(gameID string, creator engine.Player) error {
	return nil
}

// fails
func (s *stubStore) AddPendingPlayer(gameID, playerID, name string) error {
	return fmt.Errorf("that didn't work now did it")
}

func (s *stubStore) ActivateGame(gameID string) error {
	return nil
}

type fakeStore struct{}

func (s fakeStore) ActiveGames() map[string]engine.GameEngine {
	return map[string]engine.GameEngine{}
}
func (s fakeStore) InactiveGames() map[string]engine.GameEngine {
	return map[string]engine.GameEngine{}
}

func (s fakeStore) FindActiveGame(ID string) engine.GameEngine {
	// allows any game
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{Game: game.ExistingShed(game.ShedOpts{})})
	return game
}
func (s fakeStore) FindInactiveGame(ID string) engine.GameEngine {
	// allows any pending game
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{Game: game.ExistingShed(game.ShedOpts{})})
	return game
}

func (s fakeStore) FindPendingPlayer(gameID, playerID string) *protocol.Player {
	return &protocol.Player{}
}

func (s fakeStore) AddInactiveGame(game engine.GameEngine) error {
	return nil
}

func (s fakeStore) AddPendingPlayer(gameID, playerID, name string) error {
	return nil
}

func (s fakeStore) AddPlayerToGame(gameID string, player engine.Player) error {
	return nil
}

func (s fakeStore) ActivateGame(gameID string) error {
	return nil
}

func (s fakeStore) FindGame(gameID string) engine.GameEngine {
	return nil
}

func NewBasicStore() *store.InMemoryGameStore {
	return &store.InMemoryGameStore{
		Games:          map[string]engine.GameEngine{},
		PendingPlayers: map[string][]protocol.Player{},
	}
}

func mustMakeJson(t *testing.T, input interface{}) []byte {
	t.Helper()

	data, err := json.Marshal(input)
	utils.AssertNoError(t, err)

	return data
}

func newCreateGameRequest(data []byte) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, "/new", bytes.NewBuffer(data))
	return request
}

func newGetGameRequest(gameID string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/game/"+gameID, nil)
	return request
}

func newJoinGameRequest(data []byte) *http.Request {
	var request *http.Request
	if data == nil {
		request, _ = http.NewRequest(http.MethodPost, "/join", bytes.NewBuffer([]byte{}))
	} else {
		request, _ = http.NewRequest(http.MethodPost, "/join", bytes.NewBuffer(data))
	}
	return request
}

func newTestGame(t *testing.T, opts engine.GameEngineOpts) engine.GameEngine {
	t.Helper()

	game, err := engine.NewGameEngine(opts)
	if err != nil {
		t.Fatal(err)
	}
	return game
}

func newServerWithGame(game engine.GameEngine) http.Handler {
	id := game.ID()
	store := &store.InMemoryGameStore{
		Games: map[string]engine.GameEngine{
			id: game,
		},
		PendingPlayers: map[string][]protocol.Player{},
	}

	return NewServer(store)
}

// newServerWithInactiveGame returns a GameServer with an inactive game
// and some hard-coded values
func newServerWithInactiveGame(t *testing.T, ps engine.Players) (*GameServer, string) {
	t.Helper()
	gameID := "some-pending-id"
	game, err := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    gameID,
		CreatorID: "hersha-1",
		Players:   ps,
		Game:      game.ExistingShed(game.ShedOpts{}),
	})

	if err != nil {
		t.Fatal(err)
	}

	store := &store.InMemoryGameStore{
		Games: map[string]engine.GameEngine{
			gameID: game,
		},
		PendingPlayers: map[string][]protocol.Player{
			gameID: []protocol.Player{
				{
					PlayerID: "hersha-1",
					Name:     "Hersha",
				},
				{
					PlayerID: "pending-player-id",
					Name:     "Penelope",
				},
			},
		},
	}

	server := NewServer(store)

	return server, gameID
}

// newTestServerWithInactiveGame returns an httptest.Server and
// the game id of an inactive game
func newTestServerWithInactiveGame(t *testing.T, ps engine.Players, info []protocol.Player) (*httptest.Server, string) {
	store := &store.InMemoryGameStore{
		Games:          map[string]engine.GameEngine{},
		PendingPlayers: map[string][]protocol.Player{},
	}

	gameID := "some-pending-id"
	game, err := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    gameID,
		CreatorID: info[0].PlayerID,
		Players:   ps,
		Game:      game.ExistingShed(game.ShedOpts{}),
	})
	if err != nil {
		t.Fatal(err)
	}

	store.Games[gameID] = game
	store.PendingPlayers[gameID] = info

	server := httptest.NewServer(NewServer(store))

	return server, gameID
}

// newTestServer starts and returns a new server.
// The caller must call close to shut it down.
func newTestServer(str store.GameStore) *httptest.Server {
	return httptest.NewServer(NewServer(str))
}

// ASSERTIONS

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertPendingGameResponse(t *testing.T, body *bytes.Buffer, want string) {
	// probably upgrade to a jwt or something
	t.Helper()
	bodyBytes, err := ioutil.ReadAll(body)
	utils.AssertNoError(t, err)

	var got PendingGameRes
	err = json.Unmarshal(bodyBytes, &got)
	if err != nil {
		t.Fatalf("could not unmarshal json: %s", err.Error())
	}
	if got.Name != want {
		t.Errorf("got %s, want %s", got.Name, want)
	}
	if len(got.GameID) == 0 {
		t.Error("expected a game id")
	}
	if len(got.PlayerID) == 0 {
		t.Error("expected a player id")
	}
}

func mustDialWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	ws, resp, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("could not open a ws connection on %s, code %d: %s, %v", url, resp.StatusCode, body, err)
	}
	if ws == nil {
		t.Fatal("unexpected nil websocket conn")
	}

	return ws
}

func makeWSUrl(serverURL, gameID, playerID string) string {
	return "ws" + strings.Trim(serverURL, "http") +
		"/ws?gameID=" + gameID + "&playerID=" + playerID
}
