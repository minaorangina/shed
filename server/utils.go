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
	"github.com/minaorangina/shed"
	utils "github.com/minaorangina/shed/internal"
)

type stubStore struct {
	ActiveGames    map[string][]string
	PendingPlayers map[string][]string
	InactiveGames  map[string][]string
}

func (s *stubStore) FindActiveGame(ID string) shed.GameEngine {
	// allows any game
	game, _ := shed.New("", "", nil, nil)
	return game
}
func (s *stubStore) FindInactiveGame(ID string) shed.GameEngine {
	// allows any pending game
	game, _ := shed.New("", "", nil, nil)
	return game
}

// find pending player

func (s *stubStore) AddInactiveGame(gameID string, creator shed.Player) error {
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

func (s fakeStore) ActiveGames() map[string]shed.GameEngine {
	return map[string]shed.GameEngine{}
}
func (s fakeStore) InactiveGames() map[string]shed.GameEngine {
	return map[string]shed.GameEngine{}
}

func (s fakeStore) FindActiveGame(ID string) shed.GameEngine {
	// allows any game
	game, _ := shed.New("", "", nil, nil)
	return game
}
func (s fakeStore) FindInactiveGame(ID string) shed.GameEngine {
	// allows any pending game
	game, _ := shed.New("", "", nil, nil)
	return game
}

func (s fakeStore) FindPendingPlayer(gameID, playerID string) *shed.PlayerInfo {
	return &shed.PlayerInfo{}
}

func (s fakeStore) AddInactiveGame(game shed.GameEngine) error {
	return nil
}

func (s fakeStore) AddPendingPlayer(gameID, playerID, name string) error {
	return nil
}

func (s fakeStore) AddPlayerToGame(gameID string, player shed.Player) error {
	return nil
}

func (s fakeStore) ActivateGame(gameID string) error {
	return nil
}

func NewBasicStore() *shed.InMemoryGameStore {
	return &shed.InMemoryGameStore{
		ActiveGames:    map[string]shed.GameEngine{},
		InactiveGames:  map[string]shed.GameEngine{},
		PendingPlayers: map[string][]shed.PlayerInfo{},
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

func newTestGame(t *testing.T, gameID, playerID string, ps shed.Players, setupFn func(shed.GameEngine) error) shed.GameEngine {
	game, err := shed.New(gameID, playerID, ps, setupFn)
	utils.AssertNoError(t, err)

	return game
}

func newServerWithGame(game shed.GameEngine) http.Handler {
	id := game.ID()
	store := &shed.InMemoryGameStore{
		ActiveGames: map[string]shed.GameEngine{
			id: game,
		},
		InactiveGames:  map[string]shed.GameEngine{},
		PendingPlayers: map[string][]shed.PlayerInfo{},
	}

	return NewServer(store)
}

func newServerWithInactiveGame(ps shed.Players) (*GameServer, string) {
	gameID := "some-pending-id"
	game, _ := shed.New(gameID, "", ps, nil)

	store := &shed.InMemoryGameStore{
		ActiveGames: map[string]shed.GameEngine{},
		InactiveGames: map[string]shed.GameEngine{
			gameID: game,
		},
		PendingPlayers: map[string][]shed.PlayerInfo{
			gameID: []shed.PlayerInfo{{
				PlayerID: "pending-player-id",
				Name:     "Penelope",
			}},
		},
	}

	server := NewServer(store)

	return server, gameID
}

func newTestServerWithInactiveGame(ps shed.Players) (*httptest.Server, string) {
	gameID := "some-pending-id"
	game, _ := shed.New(gameID, "", ps, nil)

	store := &shed.InMemoryGameStore{
		ActiveGames: map[string]shed.GameEngine{},
		InactiveGames: map[string]shed.GameEngine{
			gameID: game,
		},
		PendingPlayers: map[string][]shed.PlayerInfo{},
	}

	server := httptest.NewServer(NewServer(store))

	return server, gameID
}

// newTestServer starts and returns a new server.
// The caller must call close to shut it down.
func newTestServer(store shed.GameStore) *httptest.Server {
	return httptest.NewServer(NewServer(store))
}

// ASSERTIONS

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertJoinGameResponse(t *testing.T, body *bytes.Buffer, ps shed.Players) {
	t.Helper()
	bodyBytes, err := ioutil.ReadAll(body)
	utils.AssertNoError(t, err)

	var got JoinGameRes
	err = json.Unmarshal(bodyBytes, &got)
	if err != nil {
		t.Fatalf("Could not unmarshal json: %s", err.Error())
	}

	if got.PlayerID == "" {
		t.Error("Expected a player id")
	}
}

func assertNewGameResponse(t *testing.T, body *bytes.Buffer, want string) {
	// probably upgrade to a jwt or something
	t.Helper()
	bodyBytes, err := ioutil.ReadAll(body)
	utils.AssertNoError(t, err)

	var got NewGameRes
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
	ws, resp, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not open a ws connection on %s, code %d: %v", url, resp.StatusCode, err)
	}

	return ws
}

func makeWSUrl(serverURL, gameID, playerID string) string {
	return "ws" + strings.Trim(serverURL, "http") +
		"/ws?game_id=" + gameID + "&player_id=" + playerID
}