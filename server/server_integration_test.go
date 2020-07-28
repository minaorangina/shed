package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerPing(t *testing.T) {
	response := httptest.NewRecorder()
	request, _ := http.NewRequest(http.MethodGet, "/", nil)

	server := NewServer()
	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusOK)
}
func TestCreateNewGame(t *testing.T) {
	t.Run("succeeds and returns expected data", func(t *testing.T) {
		data, _ := json.Marshal(NewGameReq{"Elton"})

		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/new", bytes.NewBuffer(data))

		server := NewServer()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusCreated)
		assertNewGameResponse(t, response.Body, "Elton")
	})

	t.Run("returns 400 if the player's name is missing", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/new", bytes.NewBuffer([]byte{}))
		server := NewServer()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("Does not match on GET /new", func(t *testing.T) {
		t.Skip()
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/new", nil)

		server := NewServer()
		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertNewGameResponse(t *testing.T, body *bytes.Buffer, want string) {
	t.Helper()
	bodyBytes, _ := ioutil.ReadAll(body)

	var got NewGameRes
	err := json.Unmarshal(bodyBytes, &got)
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
