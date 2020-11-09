package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/minaorangina/shed"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
)

type fakeStore struct{}

func (s fakeStore) ActiveGames() map[string]shed.GameEngine {
	return map[string]shed.GameEngine{}
}
func (s fakeStore) PendingGames() map[string]shed.GameEngine {
	return map[string]shed.GameEngine{}
}
func (s fakeStore) AddToPendingPlayers(ID string, player players.Player) error {
	return errors.New("Well that didn't work...")
}
func (s fakeStore) FindActiveGame(ID string) (shed.GameEngine, bool) {
	// allows any game
	return nil, true
}
func (s fakeStore) FindPendingGame(ID string) (shed.GameEngine, bool) {
	// allows any pending game
	return nil, true
}

func (s fakeStore) AddPendingGame(gameID string, creator players.Player) error {
	return nil
}

func (s fakeStore) ActivateGame(gameID string) error {
	return nil
}

func newBasicStore() shed.GameStore {
	return shed.NewInMemoryGameStore(nil, nil)
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

func newTestGame(t *testing.T, id string, ps players.Players, setupFn func(shed.GameEngine) error) shed.GameEngine {
	game, err := shed.New(id, ps, setupFn)
	utils.AssertNoError(t, err)

	return game
}

func newServerWithGame(game shed.GameEngine) *GameServer {
	id := game.ID()
	store := shed.NewInMemoryGameStore(
		map[string]shed.GameEngine{
			id: game,
		}, nil)

	return NewServer(store)
}

func newServerWithPendingGame(ps players.Players) (*GameServer, string) {
	gameID := "some-pending-id"
	game, _ := shed.New(gameID, ps, nil)
	store := shed.NewInMemoryGameStore(nil, map[string]shed.GameEngine{gameID: game})
	return NewServer(store), gameID
}

// ASSERTIONS

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertJoinGameResponse(t *testing.T, body *bytes.Buffer, ps players.Players) {
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
		t.Fatalf("Could not unmarshal json: %s", err.Error())
	}
	if got.Name != want {
		t.Errorf("Got %s, want %s", got.Name, want)
	}
	if len(got.GameID) == 0 {
		t.Error("Expected a game id")
	}
	if len(got.PlayerID) == 0 {
		t.Error("Expected a player id")
	}
}
