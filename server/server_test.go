package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minaorangina/shed"
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

func newJoinGameRequest(data []byte) *http.Request {
	var request *http.Request
	if data == nil {
		request, _ = http.NewRequest(http.MethodPost, "/join", bytes.NewBuffer([]byte{}))
	} else {
		request, _ = http.NewRequest(http.MethodPost, "/join", bytes.NewBuffer(data))
	}
	return request
}

func TestJoinGame(t *testing.T) {
	t.Run("POST /join returns 200 for existing game", func(t *testing.T) {
		gameID := "some-game-id"

		joiningPlayerName := "Heloise"
		data, _ := json.Marshal(JoinGameReq{gameID, joiningPlayerName})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server := newServerWithPendingGame(gameID, shed.SomePlayers())
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		assertJoinGameResponse(t, response.Body)
	})

	t.Run("POST /join returns 400 if request data missing", func(t *testing.T) {
		gameID := "some-game-id"
		ps := shed.SomePlayers()

		response := httptest.NewRecorder()
		request := newJoinGameRequest(nil)

		server := newServerWithGame(newTestGame(t, gameID, ps, nil))
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("POST /join returns 404 for an unknown game id", func(t *testing.T) {
		pendingGameID := "some-game-id"
		ps := shed.SomePlayers()

		joiningPlayerName := "Heloise"
		data, _ := json.Marshal(JoinGameReq{"fake-game-id", joiningPlayerName})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server := newServerWithPendingGame(pendingGameID, ps)
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

func newTestGame(t *testing.T, id string, ps players.Players, setupFn func(shed.GameEngine) error) shed.GameEngine {
	game, err := shed.New(id, ps, setupFn)
	testutils.AssertNoError(t, err)

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

func newServerWithPendingGame(id string, ps players.Players) *GameServer {
	store := shed.NewInMemoryGameStore(nil, map[string]players.Players{id: ps})
	return NewServer(store)
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertJoinGameResponse(t *testing.T, body *bytes.Buffer) {
	t.Helper()
	bodyBytes, err := ioutil.ReadAll(body)
	shed.AssertNoError(t, err)

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
	shed.AssertNoError(t, err)

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
