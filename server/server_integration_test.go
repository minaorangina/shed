package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestCreateAndJoinNewGame(t *testing.T) {
	// Given a request to create a new game
	name := "Ingrid"
	store := NewBasicStore()
	server := newTestServer(store)
	defer server.Close()

	data := mustMakeJson(t, NewGameReq{Name: name})
	url := server.URL + "/new"

	response, err := http.Post(url, "application/json", bytes.NewReader(data))
	utils.AssertNoError(t, err)
	defer response.Body.Close()

	// the request succeeds
	assertStatus(t, response.StatusCode, http.StatusCreated)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	utils.AssertNoError(t, err)

	// the payload contains the correct data
	var createPayload NewGameRes
	err = json.Unmarshal(bodyBytes, &createPayload)

	utils.AssertNoError(t, err)
	utils.AssertNotEmptyString(t, createPayload.GameID)
	utils.AssertNotEmptyString(t, createPayload.PlayerID)

	// an entry for the game exists in the store
	game := store.FindInactiveGame(createPayload.GameID)
	utils.AssertNotNil(t, game)

	// and a pending player is created
	utils.AssertNotNil(t, store.FindPendingPlayer(createPayload.GameID, createPayload.PlayerID))

	// Given a successful upgrade to WS for the creator
	url = makeWSUrl(server.URL, createPayload.GameID, createPayload.PlayerID)
	conn := mustDialWS(t, url)
	defer conn.Close()

	// a Player is created
	ps := game.Players()
	_, ok := ps.Find(createPayload.PlayerID)
	utils.AssertTrue(t, ok)

	// and the pending player entry is NOT removed
	// (placeholder for real auth)
	utils.AssertNotNil(t,
		store.FindPendingPlayer(createPayload.GameID, createPayload.PlayerID))

	// Given a request by a new joiner to join the game
	joinerName := "Astrid"
	data = mustMakeJson(t, JoinGameReq{GameID: createPayload.GameID, Name: joinerName})
	url = server.URL + "/join"

	response, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	utils.AssertNoError(t, err)
	defer response.Body.Close()

	// the request succeeds
	assertStatus(t, response.StatusCode, http.StatusOK)

	bodyBytes, err = ioutil.ReadAll(response.Body)
	utils.AssertNoError(t, err)

	// the payload contains the correct data
	var joinPayload JoinGameRes
	err = json.Unmarshal(bodyBytes, &joinPayload)
	utils.AssertNoError(t, err)
	utils.AssertNotEmptyString(t, joinPayload.PlayerID)

	// and a pending player is created
	utils.AssertNotNil(t,
		store.FindPendingPlayer(createPayload.GameID, joinPayload.PlayerID))

	// Given a successful upgrade to WS for the new joiner
	url = makeWSUrl(server.URL, createPayload.GameID, joinPayload.PlayerID)
	conn = mustDialWS(t, url)
	defer conn.Close()

	// a Player was created
	ps = game.Players()
	_, ok = ps.Find(joinPayload.PlayerID)
	utils.AssertTrue(t, ok)

	// and the pending player entry is NOT removed
	// (placeholder for real auth)
	utils.AssertNotNil(t,
		store.FindPendingPlayer(createPayload.GameID, joinPayload.PlayerID))
}
func TestRearrangingHand(t *testing.T) {
	// players := SomePlayers()

}
