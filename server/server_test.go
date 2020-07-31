package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	ge "github.com/minaorangina/shed/gameengine"
	testutils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
)

func TestServerPing(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	server := NewServer(nil)
	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)
}
func TestServerPOSTNewGame(t *testing.T) {
	t.Run("succeeds and returns expected data", func(t *testing.T) {
		data, _ := json.Marshal(NewGameReq{"Elton"})

		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/new", bytes.NewBuffer(data))

		server := NewServer(nil)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertNewGameResponse(t, response.Body, "Elton")
	})

	t.Run("returns 400 if the player's name is missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/new", bytes.NewBuffer([]byte{}))
		server := NewServer(nil)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Does not match on GET /new", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/new", nil)

		server := NewServer(nil)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestServerGETGame(t *testing.T) {
	t.Run("return an existing game", func(t *testing.T) {
		testID := "12u34"
		server := newServerWithGame(newTestGame(t, testID, nil, nil))

		request, _ := http.NewRequest(http.MethodGet, "/game/"+testID, nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		want := `{"game_id": "` + testID + `"}`

		bodyBytes, err := ioutil.ReadAll(response.Result().Body)
		testutils.AssertNoError(t, err)

		got := string(bodyBytes)

		testutils.AssertEqual(t, got, want)
	})

	t.Run("returns a 404 if game doesn't exist", func(t *testing.T) {
		gameID := "12u34"
		nonExistentID := "bad-game-id"
		server := newServerWithGame(newTestGame(t, gameID, nil, nil))

		request, _ := http.NewRequest(http.MethodGet, "/game/"+nonExistentID, nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutils.AssertEqual(t, response.Code, http.StatusNotFound)
	})
}

func newTestGame(t *testing.T, id string, ps []*players.Player, setupFn func(ge.GameEngine) error) ge.GameEngine {
	game, err := ge.New(id, ps, setupFn)
	testutils.AssertNoError(t, err)

	return game
}

func newServerWithGame(game ge.GameEngine) *GameServer {
	id := game.ID()
	store := &inMemoryGameStore{
		games: map[string]ge.GameEngine{
			id: game,
		},
	}

	return NewServer(store)
}
