package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

var serverTestTimeout = time.Duration(600 * time.Millisecond)

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
	var createPayload PendingGameRes
	err = json.Unmarshal(bodyBytes, &createPayload)

	utils.AssertNoError(t, err)
	utils.AssertNotEmptyString(t, createPayload.GameID)
	utils.AssertNotEmptyString(t, createPayload.PlayerID)
	utils.AssertNotEmptyString(t, createPayload.Name)
	utils.AssertTrue(t, createPayload.Admin)

	// an entry for the game exists in the store
	game := store.FindInactiveGame(createPayload.GameID)
	utils.AssertNotNil(t, game)

	// and a pending player is created
	utils.AssertNotNil(t, store.FindPendingPlayer(createPayload.GameID, createPayload.PlayerID))

	// Given a successful upgrade to WS for the creator
	url = makeWSUrl(server.URL, createPayload.GameID, createPayload.PlayerID)
	creatorConn := mustDialWS(t, url)
	defer creatorConn.Close()

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
	var joinPayload PendingGameRes
	err = json.Unmarshal(bodyBytes, &joinPayload)
	utils.AssertNoError(t, err)
	utils.AssertNotEmptyString(t, joinPayload.PlayerID)
	utils.AssertDeepEqual(t, joinPayload.Players, []protocol.Player{{createPayload.PlayerID, createPayload.Name}})
	utils.AssertEqual(t, joinPayload.Admin, false)

	// and a pending player is created
	utils.AssertNotNil(t,
		store.FindPendingPlayer(createPayload.GameID, joinPayload.PlayerID))

	// Given a successful upgrade to WS for the new joiner
	url = makeWSUrl(server.URL, createPayload.GameID, joinPayload.PlayerID)
	joinerConn := mustDialWS(t, url)
	defer joinerConn.Close()

	// a Player was created
	ps = game.Players()
	_, ok = ps.Find(joinPayload.PlayerID)
	utils.AssertTrue(t, ok) // flaky...

	// and the pending player entry is NOT removed
	// (placeholder for real auth)
	utils.AssertNotNil(t,
		store.FindPendingPlayer(createPayload.GameID, joinPayload.PlayerID))

	// and existing players are informed of the new joiner
	utils.Within(t, serverTestTimeout, func() {
		_, got, err := creatorConn.ReadMessage()
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(got) > 0)
		var payload protocol.OutboundMessage
		err = json.Unmarshal(got, &payload)
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, payload.Joiner.Name, joinerName)
	})
}

func TestStartGame(t *testing.T) {
	// Given an inactive game and players with ws connections
	creatorID, player2ID := "hersha-1", "penny-1"
	creatorName, player2Name := "Hersha", "Penelope"
	server, gameID := newTestServerWithInactiveGame(t, nil, []protocol.Player{
		{
			PlayerID: creatorID,
			Name:     creatorName,
		},
		{
			PlayerID: player2ID,
			Name:     player2Name,
		},
	})

	url := makeWSUrl(server.URL, gameID, creatorID)
	creatorConn := mustDialWS(t, url)
	defer creatorConn.Close()

	url = makeWSUrl(server.URL, gameID, player2ID)
	player2Conn := mustDialWS(t, url)
	defer player2Conn.Close()

	utils.Within(t, serverTestTimeout, func() {
		_, bytes, err := creatorConn.ReadMessage()
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(bytes) > 0)

		var data protocol.OutboundMessage
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, data.Joiner.Name, player2Name)
	})

	// When the creator sends the command to start the game
	data, err := json.Marshal(protocol.InboundMessage{PlayerID: creatorID, Command: protocol.Start})
	utils.AssertNoError(t, err)
	err = creatorConn.WriteMessage(websocket.TextMessage, data)
	utils.AssertNoError(t, err)

	// Then the start event is broadcast to all players
	utils.Within(t, serverTestTimeout, func() {
		_, bytes, err := creatorConn.ReadMessage()
		var data protocol.OutboundMessage
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, data.Command, protocol.HasStarted)

		_, bytes, err = player2Conn.ReadMessage()
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, data.Command, protocol.HasStarted)
	})

	// and it's no longer possible for players to join the game
	joinerName := "Astrid"
	data = mustMakeJson(t, JoinGameReq{GameID: gameID, Name: joinerName})
	url = server.URL + "/join"

	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	utils.AssertNoError(t, err)
	defer response.Body.Close()

	assertStatus(t, response.StatusCode, http.StatusBadRequest)
}

func TestServerNotEnoughPlayers(t *testing.T) {
	// Given a server with a game and one player with an active ws connection
	server, gameID := newTestServerWithInactiveGame(t, nil, []protocol.Player{
		{
			PlayerID: "pending-player-id",
			Name:     "Penelope",
		},
	})

	creatorID := "pending-player-id"

	url := makeWSUrl(server.URL, gameID, creatorID)
	creatorConn := mustDialWS(t, url)
	defer creatorConn.Close()

	// When the player tries to start the game
	data, err := json.Marshal(protocol.InboundMessage{
		PlayerID: creatorID,
		Command:  protocol.Start,
	})
	err = creatorConn.WriteMessage(websocket.TextMessage, data)
	utils.AssertNoError(t, err)

	// Then the player receives an error
	utils.Within(t, serverTestTimeout, func() {
		_, bytes, err := creatorConn.ReadMessage()
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(bytes) > 0)

		var data protocol.OutboundMessage
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, data.Command, protocol.Error)
	})
}

func TestServerGameStart(t *testing.T) {
	// Given a server with an inactive game
	creatorID := "player-1"
	otherPlayerID := "player-2"
	server, gameID := newTestServerWithInactiveGame(t, nil, []protocol.Player{
		{
			PlayerID: creatorID,
			Name:     "Penelope",
		},
		{
			PlayerID: otherPlayerID,
			Name:     "Wendy",
		},
	})
	// And active ws connections for each player
	creatorConn := mustDialWS(t, makeWSUrl(server.URL, gameID, creatorID))
	defer creatorConn.Close()

	p2Conn := mustDialWS(t, makeWSUrl(server.URL, gameID, otherPlayerID))
	defer p2Conn.Close()

	// When the creator starts the game
	data, err := json.Marshal(protocol.InboundMessage{
		PlayerID: creatorID,
		Command:  protocol.Start,
	})
	err = creatorConn.WriteMessage(websocket.TextMessage, data)
	utils.AssertNoError(t, err)

	// Then players are informed of the start of the game
	utils.Within(t, serverTestTimeout, func() {
		_, bytes, err := p2Conn.ReadMessage()
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(bytes) > 0)

		var data protocol.OutboundMessage
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, data.Command, protocol.HasStarted)
	})

	// And after a small delay, the game actually begins
	utils.Within(t, time.Millisecond*1500, func() {
		_, bytes, err := p2Conn.ReadMessage()
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(bytes) > 0)

		var data protocol.OutboundMessage
		err = json.Unmarshal(bytes, &data)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, data.Command, protocol.Reorg)
	})
}
