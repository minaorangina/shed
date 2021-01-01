package shed

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/minaorangina/shed/deck"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

func TestGameTurn(t *testing.T) {
	gameWithRandomPlayers := func() (int, *shed) {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(2) + 2

		players := map[string]*PlayerCards{}
		playerIDs := []string{}

		for i := 0; i < randomNumber; i++ {
			id := fmt.Sprintf("player-%d", i)
			playerIDs = append(playerIDs, id)
			players[id] = &PlayerCards{}
		}
		game := &shed{deck: deck.New(), playerCards: players, playerIDs: playerIDs}

		for i := 0; i < randomNumber; i++ {
			game.turn()
		}
		return randomNumber, game
	}

	t.Run("turns loop back through all players", func(t *testing.T) {
		numPlayers, game := gameWithRandomPlayers()

		for i := 0; i < numPlayers; i++ {
			game.turn()
		}

		utils.AssertEqual(t, game.currentTurn, 0)

		for i := 0; i < numPlayers+1; i++ {
			game.turn()
		}

		utils.AssertEqual(t, game.currentTurn, 1)
	})
}

func TestGameStageZero(t *testing.T) {
	t.Run("players asked to reorganise their hand", func(t *testing.T) {
		// Given a new game
		game := NewShed(ShedOpts{})
		// When the game starts

		err := game.Start([]string{"p1", "p2", "p3"})
		utils.AssertNoError(t, err)

		msgs, err := game.Next()

		// then players receive an instruction to reorganise their cards
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(msgs) == len(game.playerIDs))

		for _, m := range msgs {
			utils.AssertEqual(t, m.Command, protocol.Reorg)
		}

		// and the game is awaiting a response
		utils.AssertTrue(t, game.awaitingResponse)
	})

	t.Run("reorganised cards handled correctly", func(t *testing.T) {
		// Given a new game
		game := NewShed(ShedOpts{})

		// When the game has started and Next is called
		err := game.Start([]string{"p1", "p2", "p3"})
		utils.AssertNoError(t, err)
		_, err = game.Next()
		utils.AssertNoError(t, err)

		// Then the game enters the reorganisation stage
		utils.AssertEqual(t, game.stage, preGame)

		p2NewCards := somePlayerCards(3)
		p2NewCards.Unseen = game.playerCards["p2"].Unseen

		want := map[string]*PlayerCards{
			"p1": game.playerCards["p1"],
			"p2": p2NewCards,
			"p3": game.playerCards["p3"],
		}

		// and the players send their response
		msgs, err := game.ReceiveResponse([]InboundMessage{
			{
				PlayerID: "p1",
				Command:  protocol.Reorg,
				Hand:     game.playerCards["p1"].Hand,
				Seen:     game.playerCards["p1"].Seen,
			},
			{
				PlayerID: "p2",
				Command:  protocol.Reorg,
				Hand:     p2NewCards.Hand,
				Seen:     p2NewCards.Seen,
			},
			{
				PlayerID: "p3",
				Command:  protocol.Reorg,
				Hand:     game.playerCards["p3"].Hand,
				Seen:     game.playerCards["p3"].Seen,
			},
		})

		// and the response is accepted
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, msgs, []OutboundMessage(nil))

		// and players' cards are updated
		for id, c := range game.playerCards {
			utils.AssertDeepEqual(t, *c, *want[id])
		}
	})
}

func TestGameStageOne(t *testing.T) {
	t.Run("stage 1: player has legal moves", func(t *testing.T) {

		// Given a game in stage 1, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		// And a player with higher-value cards in their hand
		targetCard := deck.NewCard(deck.Nine, deck.Clubs)
		hand := []deck.Card{
			deck.NewCard(deck.Eight, deck.Hearts),
			targetCard,
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		pc := &PlayerCards{Hand: hand}

		game := NewShed(ShedOpts{
			stage:     clearDeck,
			deck:      someDeck(4),
			pile:      pile,
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		playerID := "player-1"
		oldHand := game.playerCards[playerID].Hand
		oldHandSize := len(oldHand)
		oldPileSize := len(game.pile)
		oldDeckSize := len(game.deck)

		// When the game progresses, then players are informed of the current turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))

		var moves []int

		for playerIdx, m := range msgs {
			playerID := game.playerIDs[playerIdx]

			utils.AssertEqual(t, m.PlayerID, playerID)
			utils.AssertDeepEqual(t, m.Hand, game.playerCards[playerID].Hand)
			utils.AssertDeepEqual(t, m.Seen, game.playerCards[playerID].Seen)

			if playerIdx == game.currentTurn {
				// and the current player is asked to make a choice
				utils.AssertEqual(t, m.Command, protocol.PlayHand)
				utils.AssertTrue(t, m.AwaitingResponse)
				utils.AssertTrue(t, len(m.Moves) > 0)
				moves = m.Moves
			} else {

				utils.AssertEqual(t, m.Command, protocol.Turn)
				utils.AssertEqual(t, m.AwaitingResponse, false)
			}
		}

		// And the game expects a response
		utils.AssertTrue(t, game.awaitingResponse)

		// And when player response is received
		previousTurn := game.currentTurn
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: playerID,
			Command:  protocol.PlayHand,
			Decision: []int{moves[0]},
		}})
		utils.AssertNoError(t, err)

		newHand := game.playerCards[playerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.pile)
		newDeckSize := len(game.deck)

		// Then the pile contains the selected card
		utils.AssertTrue(t, newPileSize > oldPileSize)
		utils.AssertTrue(t, containsCard(game.pile, targetCard))

		// And the deck decreases in size
		utils.AssertTrue(t, newDeckSize < oldDeckSize)

		// And the size of the player's hand remains the same
		utils.AssertTrue(t, newHandSize == oldHandSize)

		// But the cards in the player's hand changed
		utils.AssertEqual(t, reflect.DeepEqual(oldHand, newHand), false)

		// And all cards are unique
		utils.AssertTrue(t, cardsUnique(newHand))
		utils.AssertTrue(t, cardsUnique(game.pile))

		// And the game produces messages to all players, expecting no response
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))
		for _, m := range msgs {
			utils.AssertEqual(t, m.AwaitingResponse, false)
		}
		utils.AssertEqual(t, game.awaitingResponse, false)

		// And the next player is up
		utils.AssertTrue(t, game.currentTurn != previousTurn)
	})

	t.Run("stage 1: player plays multiple cards from their hand", func(t *testing.T) {
		// need to test when player has 0 or 1 card in their hand
	})

	t.Run("stage 1: player picks up pile", func(t *testing.T) {
		t.Skip()
		// Given a game with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Clubs) // Ace of Clubs

		// and a player with low-value cards in their Hand
		lowValueCards := []deck.Card{
			deck.NewCard(deck.Four, deck.Hearts),
			deck.NewCard(deck.Five, deck.Clubs),
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		game := NewShed(ShedOpts{
			stage:     clearDeck,
			deck:      someDeck(4),
			pile:      []deck.Card{highValueCard},
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": &PlayerCards{Hand: deck.Deck(lowValueCards)},
				"player-2": somePlayerCards(3),
			},
		})

		playerID := "player-1"
		oldHandSize := len(game.playerCards[playerID].Hand)
		oldPileSize := len(game.pile)
		oldDeckSize := len(game.deck)

		// when a player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		newHandSize := len(game.playerCards[playerID].Hand)
		newPileSize := len(game.pile)
		newDeckSize := len(game.deck)

		// then the current player's hand includes the cards from the pile
		utils.AssertEqual(t, newHandSize, oldHandSize+oldPileSize)
		// and the pile is now empty
		utils.AssertEqual(t, newPileSize, 0)
		// and the deck is unchanged
		utils.AssertEqual(t, newDeckSize, oldDeckSize)

		// then everyone is informed
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))

		// and the current player's OutboundMessage has the expected content
		utils.AssertTrue(t, msgs[0].AwaitingResponse)
		utils.AssertEqual(t, msgs[0].Command, protocol.NoLegalMoves)

		// and the other players' OutboundMessages have the expected content
		utils.AssertEqual(t, msgs[1].AwaitingResponse, false)
		utils.AssertEqual(t, msgs[1].Command, protocol.Turn)

		// and the current player's response is handled correctly
		previousTurn := game.currentTurn
		response, err := game.ReceiveResponse([]InboundMessage{{
			PlayerID: playerID,
			Command:  protocol.NoLegalMoves,
		}})
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, response, []OutboundMessage(nil))
		utils.AssertEqual(t, game.awaitingResponse, false)

		// and the next player is up
		utils.AssertTrue(t, game.currentTurn != previousTurn)
	})
}

func TestGameStageTwo(t *testing.T) {
	t.Run("stage 2: hand gets smaller", func(t *testing.T) {
		// Given a game in stage 2, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Six, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		// And a player with higher-value cards in their hand
		hand := []deck.Card{
			deck.NewCard(deck.Eight, deck.Hearts),
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		pc := &PlayerCards{Hand: hand}

		game := NewShed(ShedOpts{
			stage:     clearCards,
			deck:      deck.Deck{},
			pile:      pile,
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		// When a player takes their turn
		currentTurnID := game.playerIDs[game.currentTurn]
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertTrue(t, game.awaitingResponse)

		var moves []int
		for _, m := range msgs {
			if m.PlayerID == currentTurnID {
				moves = m.Moves
				break
			}
		}

		oldHandSize := len(game.playerCards[currentTurnID].Hand)
		oldSeenSize := len(game.playerCards[currentTurnID].Seen)
		oldUnseenSize := len(game.playerCards[currentTurnID].Unseen)
		previousTurnID := currentTurnID

		cardChoice := []int{moves[0]}
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: currentTurnID,
			Command:  protocol.PlayHand,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.awaitingResponse, false)

		newHandSize := len(game.playerCards[currentTurnID].Hand)
		newSeenSize := len(game.playerCards[currentTurnID].Seen)
		newUnseenSize := len(game.playerCards[currentTurnID].Unseen)
		currentTurnID = game.playerIDs[game.currentTurn]

		// Then their hand is smaller, but their remaining cards are unchanged
		utils.AssertTrue(t, newHandSize < oldHandSize)
		utils.AssertTrue(t, newSeenSize == oldSeenSize)
		utils.AssertTrue(t, newUnseenSize == oldUnseenSize)

		// And it's the next player's turn
		utils.AssertTrue(t, previousTurnID != currentTurnID)
	})
}

func TestGameReceiveResponse(t *testing.T) {
	t.Run("handles unexpected response", func(t *testing.T) {
		game := NewShed(ShedOpts{})
		err := game.Start([]string{"p1", "p2", "p3"})
		utils.AssertNoError(t, err)

		_, err = game.ReceiveResponse([]InboundMessage{{PlayerID: "p1", Command: protocol.PlayHand}})
		utils.AssertErrored(t, err)
	})

	t.Run("expects multiple responses in stage 0", func(t *testing.T) {
		game := NewShed(ShedOpts{})
		err := game.Start([]string{"p1", "p2", "p3"})
		utils.AssertNoError(t, err)

		_, err = game.Next()
		utils.AssertNoError(t, err)

		p2NewCards := somePlayerCards(3)
		p2NewCards.Unseen = game.playerCards["p2"].Unseen

		_, err = game.ReceiveResponse([]InboundMessage{
			{
				PlayerID: "p1",
				Command:  protocol.Reorg,
				Hand:     game.playerCards["p1"].Hand,
				Seen:     game.playerCards["p1"].Seen,
			},
			{
				PlayerID: "p2",
				Command:  protocol.Reorg,
				Hand:     p2NewCards.Hand,
				Seen:     p2NewCards.Seen,
			},
			{
				PlayerID: "p3",
				Command:  protocol.Reorg,
				Hand:     game.playerCards["p3"].Hand,
				Seen:     game.playerCards["p3"].Seen,
			},
		})

		utils.AssertNoError(t, err)
	})
}

func TestLegalMoves(t *testing.T) {
	type legalMoveTest struct {
		name         string
		pile, toPlay []deck.Card
		moves        []int
	}

	t.Run("four", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "four of ♣ beats two of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "four of ♠ does not beat five of ♣",
				pile: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Four, deck.Spades),
					deck.NewCard(deck.Six, deck.Spades),
					deck.NewCard(deck.Nine, deck.Hearts),
				},
				moves: []int{1, 2},
			},
			{
				name:   "four of ♥ does not beat King of ♣",
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{},
			},
			{
				name: "four of ♦ beats Seven of ♥",
				pile: []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Four, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
					deck.NewCard(deck.Ace, deck.Diamonds),
				},
				moves: []int{0, 1},
			},
			{
				name:   "four of ♣ does not beat Ace of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{},
			},
			{
				name:   "four of ♣ beats four of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "four of ♥ beats four of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("ace", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "ace of ♠ does not beat seven of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Diamonds)},
				moves:  []int{},
			},
			{
				name:   "ace of ♠ beats king of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ace of ♦ beats Four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Diamonds)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ace of ♦ beats Ace of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Diamonds)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})
	t.Run("seven", func(t *testing.T) {
		t.Run("to play", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "Seven of ♠ beats Four of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats five of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats six of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats seven of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats two of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ does not beat eight of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
				{
					name:   "seven of ♠ does not beat ace of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
		t.Run("to beat", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "four of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "five of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "six of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "Two of ♣ beats Seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "eight of ♣ does not beat seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
				{
					name:   "ace of ♣ does not beat seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
	})
	t.Run("three", func(t *testing.T) {
		t.Run("three beats anything", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "three of ♣ beats two of ♦",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
					moves:  []int{0},
				},
				{
					name:   "three of ♠ beats five of ♣",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Spades)},
					pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					moves:  []int{0},
				},
				{
					name:   "three of ♥ beats King of ♣",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Hearts)},
					pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
					moves:  []int{0},
				},
				{
					name: "three of ♦ beats Seven of ♥",
					toPlay: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.Five, deck.Clubs),
					},
					pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
					moves: []int{0, 1},
				},
				{
					name:   "three of ♣ beats Ace of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "Three of ♣ beats four of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
					moves:  []int{0},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})

		t.Run("three on the pile is ignored", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "four of ♣ does not beat six of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Six, deck.Hearts),
					},
					moves: []int{},
				},
				{
					name:   "Five of ♣ beats Four of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Four, deck.Spades),
					},
					moves: []int{0},
				},
				{
					name:   "Six of ♣ beats Five of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Five, deck.Hearts),
					},
					moves: []int{0},
				},
				{
					name:   "Seven of ♣ beats Five of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Five, deck.Spades),
					},
					moves: []int{0},
				},
				{
					name:   "Two of ♣ beats Ace of ♦",
					toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Ace, deck.Diamonds),
					},
					moves: []int{0},
				},
				{
					name:   "Eight of ♣ does not beat Nine of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Nine, deck.Spades),
					},
					moves: []int{},
				},
				{
					name:   "Jack of ♣ does not beat King of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Jack, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.King, deck.Hearts),
					},
					moves: []int{},
				},
				{
					name:   "Queen of ♣ beats Jack of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Queen, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.Jack, deck.Hearts),
					},
					moves: []int{0},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
	})
	t.Run("two beats anything; anything beats a two", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "two of ♣ beats two of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♠ beats five of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "two of ♥ beats King of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "two of ♦ beats Seven of ♥",
				toPlay: []deck.Card{
					deck.NewCard(deck.Two, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
				},
				pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				moves: []int{0, 1},
			},
			{
				name:   "Two of ♣ does beats Ace of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "two of ♥ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})
	t.Run("ten beats anything", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "ten of ♣ beats two of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♠ beats five of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♥ beats King of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "ten of ♦ beats Seven of ♥",
				toPlay: []deck.Card{
					deck.NewCard(deck.Ten, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
				},
				pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				moves: []int{0, 1},
			},
			{
				name:   "ten of ♣ does not beat Ace of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♥ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("mixture", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name: "Three, Queen and King all beat an Eight",
				pile: []deck.Card{deck.NewCard(deck.Eight, deck.Spades)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.King, deck.Diamonds),
				},
				moves: []int{0, 1, 2},
			},
			{
				name: "Three and Ten beat a King; Queen does not",
				pile: []deck.Card{
					deck.NewCard(deck.King, deck.Diamonds),
					deck.NewCard(deck.Eight, deck.Spades),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Ten, deck.Diamonds),
				},
				moves: []int{0, 2},
			},
			{
				name: "Ten beats a King hidden by a Three; Jack and Queen do not",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.King, deck.Diamonds),
					deck.NewCard(deck.Eight, deck.Spades),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Ten, deck.Diamonds),
				},
				moves: []int{2},
			},
			{
				name: "Jack, Three and Queen all beat an empty pile",
				pile: []deck.Card{},
				toPlay: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Spades),
				},
				moves: []int{0, 1, 2},
			},
			{
				name: "Three and Three beat a Queen; Jack does not",
				pile: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Three, deck.Hearts),
				},
				moves: []int{1, 2},
			},
			{
				name: "Three beats a Queen (under a Three); Seven and Jack do not",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Seven, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Hearts),
				},
				moves: []int{2},
			},
			{
				name: "Jack, Jack and Seven cannot beat a Queen (hidden by 2 Threes)",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Hearts),
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Seven, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Spades),
				},
				moves: []int{},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})
}

func TestGameStart(t *testing.T) {
	t.Run("game with no options sets up correctly", func(t *testing.T) {

		game := NewShed(ShedOpts{})
		playerIDs := []string{
			"player-1",
			"player-2",
			"player-3",
			"player-4",
		}

		err := game.Start(playerIDs)

		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(game.playerIDs) > 1)

		for _, id := range playerIDs {
			playerCards := game.playerCards[id]
			utils.AssertEqual(t, len(playerCards.Hand), 3)
			utils.AssertEqual(t, len(playerCards.Seen), 3)
			utils.AssertEqual(t, len(playerCards.Unseen), 3)
		}
	})
}

func TestGameNext(t *testing.T) {
	t.Run("game must have started", func(t *testing.T) {
		game := NewShed(ShedOpts{})
		_, err := game.Next()
		utils.AssertErrored(t, err)
	})

	t.Run("game won't progress if waiting for a response", func(t *testing.T) {
		game := NewShed(ShedOpts{awaitingResponse: true})
		err := game.Start([]string{"p1", "p2", "p3"})
		utils.AssertNoError(t, err)

		_, err = game.Next()
		utils.AssertErrored(t, err)
	})

	t.Run("new game: players reorganise cards and stage switches", func(t *testing.T) {
		// Given a new game
		game := NewShed(ShedOpts{})
		playerIDs := []string{
			"player-1",
			"player-2",
			"player-3",
			"player-4",
		}

		err := game.Start(playerIDs)
		utils.AssertNoError(t, err)

		// When Next is called
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// And players have reorganised their cards
		msgs, err = game.ReceiveResponse(reorganiseSomeCards(msgs))

		// Then the players' cards are updated in the game
		for playerIdx, m := range msgs {
			playerID := game.playerIDs[playerIdx]
			utils.AssertEqual(t, m.PlayerID, playerID)

			if playerIdx == game.currentTurn {
				utils.AssertTrue(t, m.AwaitingResponse)
				utils.AssertEqual(t, m.Command, protocol.PlayHand)
			} else {
				utils.AssertEqual(t, m.AwaitingResponse, false)
				utils.AssertEqual(t, m.Command, protocol.Turn)
			}
			utils.AssertDeepEqual(t, m.Hand, game.playerCards[playerID].Hand)
			utils.AssertDeepEqual(t, m.Seen, game.playerCards[playerID].Seen)
		}

		// And the game stage switches to stage 1
		utils.AssertEqual(t, game.stage, clearDeck)
	})

	t.Run("last card on deck: stage switches", func(t *testing.T) {
		// Given a game in stage 1
		// with a low-value card on the pile and one card left on the deck
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		game := NewShed(ShedOpts{
			stage:     clearDeck,
			deck:      someDeck(1),
			pile:      pile,
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": somePlayerCards(3),
				"player-2": somePlayerCards(3),
			},
		})

		// When the current player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)

		currentPlayerID := game.playerIDs[game.currentTurn]
		utils.AssertEqual(t, msgs[0].PlayerID, currentPlayerID)

		playerMoves := msgs[0].Moves
		utils.AssertTrue(t, len(playerMoves) > 0)

		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{playerMoves[0]}, // first possible move
		}})

		// Then the game stage switches to stage 2
		utils.AssertEqual(t, game.stage, clearCards)
	})
}

func reorganiseSomeCards(outbound []OutboundMessage) []InboundMessage {
	inbound := []InboundMessage{}
	for _, m := range outbound {
		inbound = append(inbound, InboundMessage{
			PlayerID: m.PlayerID,
			Command:  protocol.Reorg,
			// ought to shuffle really...
			Hand: m.Seen,
			Seen: m.Hand,
		})
	}

	return inbound
}

func someDeck(num int) deck.Deck {
	d := deck.New()
	d.Shuffle()
	return deck.Deck(d.Deal(num))
}

func someCards(num int) []deck.Card {
	d := someDeck(num)
	return []deck.Card(d)
}

func somePlayerCards(num int) *PlayerCards {
	return &PlayerCards{Hand: someDeck(num), Seen: someDeck(num)}
}

func containsCard(s []deck.Card, target deck.Card) bool {
	for _, c := range s {
		if c == target {
			return true
		}
	}
	return false
}
