package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minaorangina/shed"
	utils "github.com/minaorangina/shed/internal"
)

func TestCreateAndJoinNewGame(t *testing.T) {
	name := "Ingrid"
	server := NewServer(shed.NewInMemoryGameStore(nil, nil))

	data := mustMakeJson(t, NewGameReq{Name: name})

	response := httptest.NewRecorder()
	request := newCreateGameRequest(data)

	server.ServeHTTP(response, request)

	assertStatus(t, response.Code, http.StatusCreated)

	bodyBytes, err := ioutil.ReadAll(response.Result().Body)
	utils.AssertNoError(t, err)

	var payload NewGameRes
	err = json.Unmarshal(bodyBytes, &payload)
	utils.AssertNoError(t, err)

	joinerName := "Astrid"
	data = mustMakeJson(t, JoinGameReq{GameID: payload.GameID, Name: joinerName})

	joinResponse := httptest.NewRecorder()
	joinRequest := newJoinGameRequest(data)

	server.ServeHTTP(joinResponse, joinRequest)

	assertStatus(t, joinResponse.Code, http.StatusOK)

	bodyBytes, err = ioutil.ReadAll(joinResponse.Body)
	utils.AssertNoError(t, err)

	var got JoinGameRes
	err = json.Unmarshal(bodyBytes, &got)
	utils.AssertNoError(t, err)

	if got.PlayerID == "" {
		t.Error("Expected a player id")
	}
}
func TestRearrangingHand(t *testing.T) {
	// players := SomePlayers()

}
