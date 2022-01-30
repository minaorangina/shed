package game

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/minaorangina/shed/deck"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
	"github.com/stretchr/testify/assert"
)

var (
	twoPlayers   = func() []protocol.PlayerInfo { return []protocol.PlayerInfo{{PlayerID: "p1"}, {PlayerID: "p2"}} }
	threePlayers = func() []protocol.PlayerInfo {
		return []protocol.PlayerInfo{{PlayerID: "p1"}, {PlayerID: "p2"}, {PlayerID: "p3"}}
	}
	fourPlayers = func() []protocol.PlayerInfo {
		return []protocol.PlayerInfo{{PlayerID: "p1"}, {PlayerID: "p2"}, {PlayerID: "p3"}, {PlayerID: "p4"}}
	}
)

func TestGameTurn(t *testing.T) {
	gameWithRandomPlayers := func() (int, *shed) {
		rand.Seed(time.Now().UnixNano())
		randomNumberOfPlayers := rand.Intn(2) + 2

		players := map[string]*PlayerCards{}
		playerInfo := []protocol.PlayerInfo{}

		for i := 0; i < randomNumberOfPlayers; i++ {
			id := fmt.Sprintf("player-%d", i)
			playerInfo = append(playerInfo, protocol.PlayerInfo{PlayerID: id})
			players[id] = &PlayerCards{}
		}

		game := NewShed(ShedOpts{Deck: deck.New(), PlayerCards: players, PlayerInfo: playerInfo})

		return randomNumberOfPlayers, game
	}

	t.Run("turns loop back through all players", func(t *testing.T) {
		numPlayers, game := gameWithRandomPlayers()
		currentTurnIdxAtStart := game.CurrentTurnIdx

		for i := 0; i < numPlayers; i++ {
			game.turn()
		}

		utils.AssertEqual(t, game.CurrentTurnIdx, currentTurnIdxAtStart)
		utils.AssertEqual(t, game.CurrentPlayer, game.ActivePlayers[game.CurrentTurnIdx])

		assert.NotEqual(t, game.CurrentPlayer, game.NextPlayer())

		for i := 0; i < numPlayers+1; i++ {
			game.turn()
		}

		utils.AssertEqual(t, game.CurrentTurnIdx, currentTurnIdxAtStart+1)
		utils.AssertEqual(t, game.CurrentPlayer.PlayerID, "player-1")
	})
}

func TestNewShed(t *testing.T) {
	t.Run("game with no options sets up correctly", func(t *testing.T) {
		t.Log("Given a new game")
		game := NewShed(ShedOpts{})
		playerInfo := fourPlayers()

		t.Log("When the game starts")
		err := game.Start(playerInfo)

		utils.AssertNoError(t, err)

		t.Log("Then the players are initiated correctly")
		utils.AssertTrue(t, len(game.PlayerInfo) > 1)
		utils.AssertTrue(t, len(game.ActivePlayers) == len(game.PlayerInfo))
		utils.AssertNotEmptyString(t, game.CurrentPlayer.PlayerID)

		t.Log("And the game is in the correct gameplay state")
		assert.Equal(t, gameStarted, game.gamePlay)
		assert.False(t, game.GameOver())

		t.Log("And players' cards are set correctly")
		for _, p := range game.PlayerCards {
			utils.AssertEqual(t, len(p.UnseenVisibility), 3)
		}

		for _, info := range playerInfo {
			id := info.PlayerID
			playerCards := game.PlayerCards[id]
			utils.AssertEqual(t, len(playerCards.Hand), 3)
			utils.AssertEqual(t, len(playerCards.Seen), 3)
			utils.AssertEqual(t, len(playerCards.Unseen), 3)
		}
	})

	t.Run("existing game must have players", func(t *testing.T) {
		tf := func() {
			NewShed(ShedOpts{Pile: someDeck(6)})
		}
		assert.Panics(t, tf)
	})

	// t.Run("game with options sets up correctly", func(t *testing.T) {

	// })
}

func TestGameNext(t *testing.T) {
	t.Run("game must have started", func(t *testing.T) {
		game := NewShed(ShedOpts{})
		_, err := game.Next()
		utils.AssertErrored(t, err)
	})

	t.Run("game won't progress if waiting for a response", func(t *testing.T) {
		game := NewShed(ShedOpts{ExpectedCommand: protocol.PlaySeen})
		err := game.Start(threePlayers())
		utils.AssertNoError(t, err)

		_, err = game.Next()
		utils.AssertErrored(t, err)
	})

	// test contents of messages
	t.Run("new game: players reorganise cards and stage switches", func(t *testing.T) {
		// Given a new game
		game := NewShed(ShedOpts{})

		err := game.Start(fourPlayers())
		utils.AssertNoError(t, err)

		// When Next is called
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// And players have reorganised their cards
		msgs, err = game.ReceiveResponse(reorganiseSomeCards(msgs))

		// Then the players' cards are updated in the game
		for playerIdx, m := range msgs {
			playerID := game.PlayerInfo[playerIdx].PlayerID
			utils.AssertEqual(t, m.PlayerID, playerID)

			if playerIdx == game.CurrentTurnIdx {
				utils.AssertTrue(t, m.ShouldRespond)
				utils.AssertEqual(t, m.Command, protocol.PlayHand)
			} else {
				utils.AssertEqual(t, m.ShouldRespond, protocol.Null)
				utils.AssertEqual(t, m.Command, protocol.Turn)
			}
			utils.AssertDeepEqual(t, m.Hand, game.PlayerCards[playerID].Hand)
			utils.AssertDeepEqual(t, m.Seen, game.PlayerCards[playerID].Seen)
		}

		// And the game stage switches to stage 1
		utils.AssertEqual(t, game.Stage, clearDeck)
	})

	t.Run("last card on Deck: stage switches", func(t *testing.T) {
		t.SkipNow()
		// Given a game in stage 1
		// with a low-value card on the pile and one card left on the deck
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		game := NewShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(1),
			Pile:          pile,
			PlayerInfo:    twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": somePlayerCards(3),
				"p2": somePlayerCards(3),
			},
		})

		// When the current player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		utils.AssertEqual(t, msgs[0].PlayerID, game.CurrentPlayer.PlayerID)

		playerMoves := msgs[0].Moves
		utils.AssertTrue(t, len(playerMoves) > 0)

		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{playerMoves[0]}, // first possible move
		}})

		// Then the game expects an ack
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.ReplenishHand)

		// And when the game receives the ack
		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.ReplenishHand,
		}})

		// Then the game stage switches to stage 2
		utils.AssertEqual(t, game.Stage, clearCards)
	})
}

func TestGameReceiveResponse(t *testing.T) {
	t.Run("will fail if game not started", func(t *testing.T) {
		game := NewShed(ShedOpts{
			Stage:           1,
			ExpectedCommand: protocol.PlayHand,
			CurrentPlayer:   threePlayers()[0],
			Deck:            []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
			PlayerCards: map[string]*PlayerCards{
				"p1": {Hand: someCards(3)},
				"p2": {Hand: someCards(3)},
				"p3": {Hand: someCards(3)},
			},
		})
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)
		playerID := game.CurrentPlayer.PlayerID
		_, err := game.ReceiveResponse([]protocol.InboundMessage{{PlayerID: playerID, Command: protocol.PlayHand}})
		assert.ErrorIs(t, err, ErrGameNotStarted)
	})

	t.Run("handles unexpected response", func(t *testing.T) {
		t.SkipNow()
		game := NewShed(ShedOpts{Stage: 1, CurrentPlayer: threePlayers()[0]})
		err := game.Start(threePlayers())
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		_, err = game.ReceiveResponse([]protocol.InboundMessage{{PlayerID: "p1", Command: protocol.PlayHand}})
		utils.AssertErrored(t, err)
	})

	t.Run("handles response from wrong player", func(t *testing.T) {
		t.Skip()
		game := NewShed(ShedOpts{
			Stage:           1,
			ExpectedCommand: protocol.PlayHand,
			CurrentPlayer:   threePlayers()[0],
			Deck:            []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
			PlayerCards: map[string]*PlayerCards{
				"p1": {Hand: someCards(3)},
				"p2": {Hand: someCards(3)},
				"p3": {Hand: someCards(3)},
			},
		})
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		msgs, err := game.ReceiveResponse([]protocol.InboundMessage{{PlayerID: "p2", Command: protocol.PlayHand}})
		utils.AssertErrored(t, err)
		utils.AssertContains(t, err.Error(), "unexpected message from player")

		// Player is sent an error message
		utils.AssertEqual(t, len(msgs), 1)
		utils.AssertEqual(t, msgs[0].PlayerID, "p2")
	})

	t.Run("handles response with incorrect command", func(t *testing.T) {
		t.SkipNow()
		game := NewShed(ShedOpts{
			Stage:           1,
			ExpectedCommand: protocol.PlayHand,
			CurrentPlayer:   threePlayers()[0],
			Deck:            []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
			PlayerCards: map[string]*PlayerCards{
				"p1": {Hand: someCards(3)},
				"p2": {Hand: someCards(3)},
				"p3": {Hand: someCards(3)},
			},
		})
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		msgs, err := game.ReceiveResponse([]protocol.InboundMessage{{PlayerID: "p1", Command: protocol.PlayUnseen}})
		utils.AssertErrored(t, err)
		utils.AssertContains(t, err.Error(), "unexpected command")

		// Player is sent an error message
		utils.AssertEqual(t, len(msgs), 1)
		utils.AssertEqual(t, msgs[0].PlayerID, game.CurrentPlayer.PlayerID)
	})

	t.Run("expects multiple responses in stage 0", func(t *testing.T) {
		game := NewShed(ShedOpts{})
		err := game.Start(threePlayers())
		utils.AssertNoError(t, err)

		_, err = game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Reorg)

		p2NewCards := somePlayerCards(3)
		p2NewCards.Unseen = game.PlayerCards["p2"].Unseen

		_, err = game.ReceiveResponse([]protocol.InboundMessage{
			{
				PlayerID: "p1",
				Command:  protocol.Reorg,
				Decision: []int{2, 3, 4},
			},
			{
				PlayerID: "p2",
				Command:  protocol.Reorg,
				Decision: []int{0, 1, 5},
			},
			{
				PlayerID: "p3",
				Command:  protocol.Reorg,
				Decision: []int{1, 3, 4},
			},
		})

		utils.AssertNoError(t, err)
	})

	t.Run("expects one card choice in stage 2 unseen", func(t *testing.T) {
		t.SkipNow()
		game := NewShed(ShedOpts{
			Stage:           clearCards,
			ExpectedCommand: protocol.PlayUnseen,
			CurrentPlayer:   threePlayers()[0],
			Deck:            []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
			PlayerCards: map[string]*PlayerCards{
				"p1": {Unseen: someCards(3)},
				"p2": {Seen: someCards(3)},
				"p3": {Seen: someCards(3)},
			},
		})
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayUnseen)

		msgs, err := game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: "p1",
			Decision: []int{0, 1},
			Command:  protocol.PlayUnseen,
		}})

		utils.AssertErrored(t, err)
		utils.AssertEqual(t, err, ErrPlayOneCard)
		utils.AssertEqual(t, len(msgs), 1)
		utils.AssertEqual(t, msgs[0].Command, protocol.Error)
		utils.AssertTrue(t, strings.Contains(msgs[0].Message, ErrPlayOneCard.Error()))
	})

	t.Run("sends error messages if game in bad state", func(t *testing.T) {
		game := NewShed(ShedOpts{
			Stage:           clearDeck,
			ExpectedCommand: protocol.PlayUnseen, // this is impossible in clearDeck stage
			CurrentPlayer:   threePlayers()[0],
			Deck:            []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
			PlayerCards: map[string]*PlayerCards{
				"p1": {Unseen: someCards(3)},
				"p2": {Seen: someCards(3)},
				"p3": {Seen: someCards(3)},
			},
		})
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayUnseen)

		msgs, err := game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: "p1",
			Decision: []int{0, 1},
			Command:  protocol.PlayUnseen,
		}})

		utils.AssertErrored(t, err)
		utils.AssertEqual(t, err, ErrInvalidGameState)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
		for _, m := range msgs {
			utils.AssertEqual(t, m.Command, protocol.Error)
			utils.AssertTrue(t, strings.Contains(m.Message, ErrInvalidGameState.Error()))
		}
	})

}

func TestGameBurn(t *testing.T) {
	ps1 := twoPlayers()
	ps2 := twoPlayers()
	tt := []struct {
		name            string
		opts            ShedOpts
		decision        []int
		expectedCommand protocol.Cmd
	}{
		{
			name:            "ten on a four",
			decision:        []int{0},
			expectedCommand: protocol.PlayHand,
			opts: ShedOpts{
				Stage: clearCards,
				Deck:  []deck.Card{},
				Pile: []deck.Card{
					deck.NewCard(deck.Six, deck.Diamonds),
					deck.NewCard(deck.Six, deck.Spades),
					deck.NewCard(deck.Eight, deck.Hearts),
					deck.NewCard(deck.Nine, deck.Hearts),
					deck.NewCard(deck.Two, deck.Spades),
					deck.NewCard(deck.Seven, deck.Clubs),
					deck.NewCard(deck.Four, deck.Spades),
				},
				PlayerInfo:    ps1,
				CurrentPlayer: ps1[1],
				PlayerCards: map[string]*PlayerCards{
					"p1": somePlayerCards(3),
					"p2": {
						Hand: []deck.Card{
							deck.NewCard(deck.Ten, deck.Hearts),
							deck.NewCard(deck.Ace, deck.Hearts),
							deck.NewCard(deck.Queen, deck.Clubs),
						},
					},
				},
			},
		},
		{
			name:            "play last six",
			decision:        []int{1},
			expectedCommand: protocol.PlayHand,
			opts: ShedOpts{
				Stage: clearDeck,
				Deck:  someDeck(4),
				Pile: []deck.Card{
					deck.NewCard(deck.Four, deck.Hearts),
					deck.NewCard(deck.Two, deck.Spades),
					deck.NewCard(deck.Six, deck.Hearts),
					deck.NewCard(deck.Six, deck.Spades),
					deck.NewCard(deck.Six, deck.Clubs),
				},
				PlayerInfo:    ps2,
				CurrentPlayer: ps2[0],
				PlayerCards: map[string]*PlayerCards{
					"p1": {
						Hand: []deck.Card{
							deck.NewCard(deck.Eight, deck.Hearts),
							deck.NewCard(deck.Six, deck.Diamonds),
							deck.NewCard(deck.Five, deck.Diamonds),
						},
					},
					"p2": somePlayerCards(3),
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Given a game in stage
			game := NewShed(tc.opts)

			msgs, err := game.Next()
			utils.AssertNoError(t, err)
			utils.AssertEqual(t, game.AwaitingResponse(), tc.expectedCommand)

			checkNextMessages(t, msgs, tc.expectedCommand, game)

			// When the player plays their move
			msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
				PlayerID: game.CurrentPlayer.PlayerID,
				Command:  tc.expectedCommand,
				Decision: tc.decision, // target card is the second one
			}})
			utils.AssertNoError(t, err)

			// Then the game sends burn messages to all players
			// expecting a response only from the current player
			utils.AssertEqual(t, game.AwaitingResponse(), protocol.Burn)
			utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
			checkBurnMessages(t, msgs, game)

			// But the deck has not been burned yet
			utils.AssertTrue(t, len(game.Pile) > 0)

			// And when the current player acks
			previousPlayerID := game.CurrentPlayer.PlayerID
			msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
				PlayerID: game.CurrentPlayer.PlayerID,
				Command:  protocol.Burn,
			}})
			utils.AssertNoError(t, err)

			// Then the selected card has been burned along with the pile
			utils.AssertEqual(t, len(game.Pile), 0)
			// utils.AssertEqual(t, containsCard(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand, targetCard), false)

			if len(game.Deck) == 0 {
				utils.AssertEqual(t, len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand), 2)
			}

			// And the current player gets another turn
			utils.AssertEqual(t, game.CurrentPlayer.PlayerID, previousPlayerID)
		})
	}

	t.Run("Burn on UnseenSuccess", func(t *testing.T) {
		// Given a game
		ps3 := twoPlayers()
		game := NewShed(ShedOpts{
			Stage:         clearCards,
			Deck:          []deck.Card{},
			Pile:          []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
			PlayerInfo:    ps3,
			CurrentPlayer: ps3[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": NewPlayerCards(nil, nil, []deck.Card{
					deck.NewCard(deck.Ten, deck.Diamonds),
					deck.NewCard(deck.Seven, deck.Diamonds),
				}, nil),
				"p2": somePlayerCards(3),
			},
		})

		expectedCommand := protocol.PlayUnseen
		decision := []int{0}

		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), expectedCommand)

		checkNextMessages(t, msgs, expectedCommand, game)

		// When the player plays their move
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  expectedCommand,
			Decision: decision, // target card is the second one
		}})
		utils.AssertNoError(t, err)

		// Then the game sends UnseenSuccess to everyone
		// expecting a response only from the current player
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.UnseenSuccess)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))

		// And when the player acks
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.UnseenSuccess,
		}})
		utils.AssertNoError(t, err)

		// Then the game sends burn messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Burn)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
		checkBurnMessages(t, msgs, game)

		// But the deck has not been burned yet
		utils.AssertTrue(t, len(game.Pile) > 0)

		// And when the current player acks
		previousPlayerID := game.CurrentPlayer.PlayerID
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.Burn,
		}})
		utils.AssertNoError(t, err)

		// Then the deck has been burned
		utils.AssertEqual(t, len(game.Deck), 0)

		// Then the current player gets another turn
		utils.AssertEqual(t, game.CurrentPlayer.PlayerID, previousPlayerID)
	})
}

func checkBaseMessage(t *testing.T, m protocol.OutboundMessage, game *shed) {
	t.Helper()

	publicUnseen := game.mapUnseenToPublicUnseen(m.PlayerID)

	utils.AssertNotEmptyString(t, m.PlayerID)
	utils.AssertDeepEqual(t, m.CurrentTurn, game.CurrentPlayer)
	utils.AssertDeepEqual(t, m.Hand, game.PlayerCards[m.PlayerID].Hand)
	utils.AssertDeepEqual(t, m.Seen, game.PlayerCards[m.PlayerID].Seen)
	utils.AssertDeepEqual(t, m.Unseen, publicUnseen)
	utils.AssertDeepEqual(t, m.Pile, game.Pile)
	utils.AssertEqual(t, m.DeckCount, len(game.Deck))
}

func checkNextMessages(t *testing.T, msgs []protocol.OutboundMessage, cmd protocol.Cmd, game *shed) {
	t.Helper()

	for _, m := range msgs {
		checkBaseMessage(t, m, game)

		if m.PlayerID == game.CurrentPlayer.PlayerID {
			// and the current player is asked to make a choice
			utils.AssertEqual(t, m.Command, cmd)
			utils.AssertTrue(t, m.ShouldRespond)
			utils.AssertTrue(t, len(m.Moves) > 0)
		} else {
			cmdForOtherPlayers := protocol.Turn
			if cmd == protocol.ReplenishHand {
				cmdForOtherPlayers = protocol.EndOfTurn
			}
			utils.AssertEqual(t, m.Command, cmdForOtherPlayers)
			utils.AssertEqual(t, m.ShouldRespond, false)
		}
	}
}

func checkReceiveResponseMessages(t *testing.T, msgs []protocol.OutboundMessage, currentPlayerCmd protocol.Cmd, game *shed) {
	t.Helper()

	for _, m := range msgs {
		checkBaseMessage(t, m, game)

		if m.PlayerID == game.CurrentPlayer.PlayerID {
			// and the current player is asked to make a choice
			utils.AssertEqual(t, m.Command, currentPlayerCmd)
			utils.AssertTrue(t, m.ShouldRespond)
		} else {
			utils.AssertEqual(t, m.Command, protocol.EndOfTurn)
			utils.AssertEqual(t, m.ShouldRespond, false)
		}
	}
}

func checkPlayerFinishedMessages(t *testing.T, msgs []protocol.OutboundMessage, game *shed) {
	t.Helper()

	for _, m := range msgs {
		checkBaseMessage(t, m, game)
		utils.AssertEqual(t, m.Command, protocol.PlayerFinished)

		if m.PlayerID == game.CurrentPlayer.PlayerID {
			utils.AssertTrue(t, m.ShouldRespond)
			utils.AssertEqual(t, len(m.Unseen), 0)
		} else {
			utils.AssertEqual(t, m.ShouldRespond, false)
		}
	}
}

func checkGameOverMessages(t *testing.T, msgs []protocol.OutboundMessage, game *shed) {
	t.Helper()

	for _, m := range msgs {
		checkBaseMessage(t, m, game)
		utils.AssertEqual(t, m.Command, protocol.GameOver)
		utils.AssertEqual(t, m.ShouldRespond, false)
		utils.AssertTrue(t, len(game.FinishedPlayers) == len(game.PlayerInfo))
		utils.AssertDeepEqual(t, m.FinishedPlayers, game.FinishedPlayers)
	}
}

func checkBurnMessages(t *testing.T, msgs []protocol.OutboundMessage, game *shed) {
	t.Helper()

	for _, m := range msgs {
		checkBaseMessage(t, m, game)
		utils.AssertEqual(t, m.Command, protocol.Burn)
		if m.PlayerID == game.CurrentPlayer.PlayerID {
			utils.AssertTrue(t, m.ShouldRespond)
		}
	}
}

func reorganiseSomeCards(outbound []protocol.OutboundMessage) []protocol.InboundMessage {
	inbound := []protocol.InboundMessage{}
	for _, m := range outbound {
		inbound = append(inbound, protocol.InboundMessage{
			PlayerID: m.PlayerID,
			Command:  protocol.Reorg,
			// ought to shuffle really...
			Decision: []int{2, 4, 5},
		})
	}

	return inbound
}

func getMoves(msgs []protocol.OutboundMessage, currentPlayerID string) []int {
	var moves []int
	for _, m := range msgs {
		if m.PlayerID == currentPlayerID {
			moves = m.Moves
		}
	}
	return moves
}

func getNonBurnMove(cards []deck.Card, moves []int) int {
	for _, v := range moves {
		if cards[v].Rank != deck.Ten {
			return v
		}
	}
	panic(fmt.Sprintf("all cards are a Ten: %v", cards))
}
