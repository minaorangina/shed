package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed"
	utils "github.com/minaorangina/shed/internal"
)

func TestServerPing(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	server := NewServer(NewBasicStore())
	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	utils.AssertNoError(t, err)
	utils.AssertTrue(t, strings.Contains(string(bodyBytes), "<!DOCTYPE html>"))
}

func TestStatic(t *testing.T) {
	t.Run("fetches css", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/static/index.css", nil)

		server := NewServer(NewBasicStore())
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		bodyBytes, err := ioutil.ReadAll(response.Body)
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), ".form-error"))
	})
}
func TestServerPOSTNewGame(t *testing.T) {
	t.Run("succeeds and returns expected data", func(t *testing.T) {
		data := mustMakeJson(t, NewGameReq{"Elton"})

		response := httptest.NewRecorder()
		request := newCreateGameRequest(data)

		server := NewServer(NewBasicStore())
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertPendingGameResponse(t, response.Body, "Elton")
	})

	t.Run("returns 400 if the player's name is missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := newCreateGameRequest([]byte{})

		server := NewServer(NewBasicStore())
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

func TestGETGameWaitingRoom(t *testing.T) {
	t.Run("fails for bad credentials", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/waiting-room", nil)

		server := NewServer(NewBasicStore())
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("game creator sees a special admin view", func(t *testing.T) {
		gameID, creatorID := "some-game-id", "i-am-the-creator"

		url := "/waiting-room?game_id=" + gameID + "&player_id=" + creatorID

		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, url, nil)

		store := NewBasicStore()
		game, err := shed.NewGameEngine(shed.GameEngineOpts{
			GameID:    gameID,
			CreatorID: creatorID,
			Game:      shed.NewShed(),
		})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, game)

		store.Games[gameID] = game
		store.PendingPlayers[gameID] = []shed.PlayerInfo{{PlayerID: creatorID, Name: "Horatio"}}

		server := NewServer(store)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		bodyBytes, err := ioutil.ReadAll(response.Body)
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "<!DOCTYPE html>"))
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "</html>"))
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "created"))
	})

	t.Run("game joiners see standard view", func(t *testing.T) {
		gameID, playerID := "some-game-id", "i-am-some-rando"

		url := "/waiting-room?game_id=" + gameID + "&player_id=" + playerID

		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, url, nil)

		store := NewBasicStore()
		game, err := shed.NewGameEngine(shed.GameEngineOpts{
			GameID:    gameID,
			CreatorID: "i-am-the-creator",
			Game:      shed.NewShed(),
		})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, game)

		store.Games[gameID] = game
		store.PendingPlayers[gameID] = []shed.PlayerInfo{{PlayerID: playerID, Name: "Horatio"}}

		server := NewServer(store)
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)
		bodyBytes, err := ioutil.ReadAll(response.Body)
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "<!DOCTYPE html>"))
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "</html>"))
		utils.AssertTrue(t, strings.Contains(string(bodyBytes), "Joined"))
	})
}

func TestJoinGame(t *testing.T) {
	t.Run("POST /join returns 200 for existing game", func(t *testing.T) {
		server, pendingID := newServerWithInactiveGame(t, shed.SomePlayers())

		joiningPlayerName := "Heloise"
		data := mustMakeJson(t, JoinGameReq{pendingID, joiningPlayerName})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusOK)

		bodyBytes, err := ioutil.ReadAll(response.Body)
		utils.AssertNoError(t, err)

		var got PendingGameRes
		err = json.Unmarshal(bodyBytes, &got)
		if err != nil {
			t.Fatalf("Could not unmarshal json: %s", err.Error())
		}
		if got.PlayerID == "" {
			t.Error("Expected a player id")
		}
		if got.GameID == "" {
			t.Error("Expected a game id")
		}
		if got.Name == "" {
			t.Error("Expected a player name")
		}
	})

	t.Run("POST /join returns 400 if request data missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := newJoinGameRequest(nil)

		game := newTestGame(t, shed.GameEngineOpts{GameID: "some-game-id", Players: shed.SomePlayers(), Game: shed.NewShed()})
		server := newServerWithGame(game)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("POST /join returns 400 for an unknown game id", func(t *testing.T) {
		server, _ := newServerWithInactiveGame(t, shed.SomePlayers())

		data := mustMakeJson(t, JoinGameReq{"some-game-id", "Heloise"})

		response := httptest.NewRecorder()
		request := newJoinGameRequest(data)

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("POST /join returns 500 if joining fails", func(t *testing.T) {
		t.Skip("Will reinstate this test when db is being used")

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
		server := newServerWithGame(newTestGame(t, shed.GameEngineOpts{
			GameID:    testID,
			PlayState: shed.InProgress,
			Game:      shed.NewShed(),
		}))

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
		server, pendingID := newServerWithInactiveGame(t, shed.SomePlayers())

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
		server := newServerWithGame(newTestGame(t, shed.GameEngineOpts{GameID: gameID, Game: shed.NewShed()}))

		request := newGetGameRequest(nonExistentID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		utils.AssertEqual(t, response.Code, http.StatusNotFound)
	})
}

func TestWS(t *testing.T) {
	t.Run("Handles missing game details", func(t *testing.T) {
		server := httptest.NewServer(NewServer(NewBasicStore()))

		_, _, err := websocket.DefaultDialer.Dial("ws"+strings.Trim(server.URL, "http")+"/ws", nil)
		utils.AssertErrored(t, err)
	})

	t.Run("Rejects if pending game doesn't exist", func(t *testing.T) {
		name, playerID := "Delilah", "delilah1"
		p := shed.APlayer(playerID, name)
		ps := shed.NewPlayers(p)

		server, _ := newTestServerWithInactiveGame(t, ps)
		defer server.Close()

		wsURL := "ws" + strings.Trim(server.URL, "http") +
			"/ws?game_id=unknowngamelol&player_id=unknownhooman"

		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		utils.AssertErrored(t, err)
		utils.AssertEqual(t, resp.StatusCode, http.StatusBadRequest)
	})

	t.Run("Successfully connects", func(t *testing.T) {
		gameID := "this-is-a-game-id"
		name, playerID := "Delilah", "delilah1"

		game := newTestGame(t, shed.GameEngineOpts{GameID: gameID, CreatorID: playerID, Game: shed.NewShed()})

		store := NewBasicStore()
		store.AddInactiveGame(game)
		store.AddPendingPlayer(gameID, playerID, name)

		server := newTestServer(store)
		defer server.Close()

		wsURL := "ws" + strings.Trim(server.URL, "http") +
			"/ws?game_id=" + gameID + "&player_id=" + playerID

		ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		utils.AssertEqual(t, resp.StatusCode, http.StatusSwitchingProtocols)
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, ws)
	})
}
