package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
)

func TestServerPing(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	server := NewServer(newBasicStore())
	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	utils.AssertNoError(t, err)
	utils.AssertTrue(t, strings.Contains(string(bodyBytes), "<!DOCTYPE html>"))
}
func TestServerPOSTNewGame(t *testing.T) {
	t.Run("succeeds and returns expected data", func(t *testing.T) {
		data := mustMakeJson(t, NewGameReq{"Elton"})

		response := httptest.NewRecorder()
		request := newCreateGameRequest(data)

		server := NewServer(newBasicStore())
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertNewGameResponse(t, response.Body, "Elton")
	})

	t.Run("returns 400 if the player's name is missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := newCreateGameRequest([]byte{})

		server := NewServer(newBasicStore())
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

func TestGETGameCreated(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/created", nil)

	server := NewServer(newBasicStore())
	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)
	bodyBytes, err := ioutil.ReadAll(response.Body)
	utils.AssertNoError(t, err)
	utils.AssertTrue(t, strings.Contains(string(bodyBytes), "<!DOCTYPE html>"))
}

func TestJoinGame(t *testing.T) {
	t.Run("POST /join returns 200 for existing game", func(t *testing.T) {
		server, pendingID := newServerWithPendingGame(players.SomePlayers())

		joiningPlayerName := "Heloise"
		data := mustMakeJson(t, JoinGameReq{pendingID, joiningPlayerName})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		bodyBytes, err := ioutil.ReadAll(response.Body)
		utils.AssertNoError(t, err)

		var got JoinGameRes
		err = json.Unmarshal(bodyBytes, &got)
		if err != nil {
			t.Fatalf("Could not unmarshal json: %s", err.Error())
		}

		if got.PlayerID == "" {
			t.Error("Expected a player id")
		}
	})

	t.Run("POST /join returns 400 if request data missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := newJoinGameRequest(nil)

		game := newTestGame(t, "some-game-id", players.SomePlayers(), nil)
		server := newServerWithGame(game)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("POST /join returns 404 for an unknown game id", func(t *testing.T) {
		server, _ := newServerWithPendingGame(players.SomePlayers())

		data := mustMakeJson(t, JoinGameReq{"some-game-id", "Heloise"})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("POST /join returns 500 if joining fails", func(t *testing.T) {
		data := mustMakeJson(t, JoinGameReq{"some-game-id", "Heloise"})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server := NewServer(fakeStore{})

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func TestServerGETGame(t *testing.T) {
	t.Run("returns an existing active game", func(t *testing.T) {
		testID := "12u34"
		server := newServerWithGame(newTestGame(t, testID, nil, nil))

		request := newGetGameRequest(testID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		want := GetGameRes{Status: "active", GameID: testID}

		bodyBytes, err := ioutil.ReadAll(response.Result().Body)
		utils.AssertNoError(t, err)

		var got GetGameRes
		err = json.Unmarshal(bodyBytes, &got)
		utils.AssertNoError(t, err)

		assertStatus(t, response.Code, http.StatusOK)
		utils.AssertEqual(t, got, want)
	})

	t.Run("returns an existing pending game", func(t *testing.T) {
		server, pendingID := newServerWithPendingGame(players.SomePlayers())

		request := newGetGameRequest(pendingID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		want := GetGameRes{Status: "pending", GameID: pendingID}

		bodyBytes, err := ioutil.ReadAll(response.Result().Body)
		utils.AssertNoError(t, err)

		var got GetGameRes
		err = json.Unmarshal(bodyBytes, &got)
		utils.AssertNoError(t, err)

		assertStatus(t, response.Code, http.StatusOK)
		utils.AssertEqual(t, got, want)
	})

	t.Run("returns a 404 if game doesn't exist", func(t *testing.T) {
		gameID := "12u34"
		nonExistentID := "bad-game-id"
		server := newServerWithGame(newTestGame(t, gameID, nil, nil))

		request := newGetGameRequest(nonExistentID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		utils.AssertEqual(t, response.Code, http.StatusNotFound)
	})
}
