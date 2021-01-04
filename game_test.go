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

		utils.AssertEqual(t, game.currentTurnIdx, 0)

		for i := 0; i < numPlayers+1; i++ {
			game.turn()
		}

		utils.AssertEqual(t, game.currentTurnIdx, 1)
		utils.AssertEqual(t, game.currentPlayerID, "player-1")
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
			stage:           clearDeck,
			deck:            someDeck(4),
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		oldHand := game.playerCards[game.currentPlayerID].Hand
		oldHandSize, oldPileSize, oldDeckSize := len(oldHand), len(game.pile), len(game.deck)

		// When the game progresses, then players are informed of the current turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))

		checkNextMessages(t, msgs, protocol.PlayHand, game)
		moves := getMoves(msgs, game.currentPlayerID)

		// And the game expects a response
		utils.AssertTrue(t, game.awaitingResponse)

		// And when player response is received
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{moves[1]}, // target card is the second one
		}})
		utils.AssertNoError(t, err)

		newHand := game.playerCards[game.currentPlayerID].Hand
		newHandSize, newPileSize, newDeckSize := len(newHand), len(game.pile), len(game.deck)

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
		utils.AssertTrue(t, cardsUnique(newHand)) // this fails sometimes
		utils.AssertTrue(t, cardsUnique(game.pile))

		// And the game produces messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))
		checkReceiveResponseMessages(t, msgs, protocol.ReplenishHand, game)
		utils.AssertTrue(t, game.awaitingResponse)

		// And when the current player acks and releases their turn
		previousPlayerID := game.currentPlayerID
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.ReplenishHand,
		}})
		utils.AssertNoError(t, err)

		// And the next player is up
		utils.AssertTrue(t, game.currentPlayerID != previousPlayerID)
	})

	t.Run("stage 1: player plays multiple cards from their hand", func(t *testing.T) {
		// Given a game in stage 1
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}
		targetCards := []deck.Card{
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Nine, deck.Diamonds),
		}
		// And a player with two cards of the same value in their hand
		pc := &PlayerCards{Hand: append(targetCards, deck.NewCard(deck.Eight, deck.Hearts))}

		game := NewShed(ShedOpts{
			stage:           clearDeck,
			deck:            someDeck(4),
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		oldHand := game.playerCards[game.currentPlayerID].Hand
		oldHandSize := len(oldHand)
		oldPileSize := len(game.pile)
		oldDeckSize := len(game.deck)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, game.currentTurnIdx, 0)
		utils.AssertTrue(t, game.currentPlayerID != "")

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		// And chooses to play two of the same card
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0, 1},
		}})
		utils.AssertNoError(t, err)
		newHand := game.playerCards[game.currentPlayerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.pile)
		newDeckSize := len(game.deck)

		// Then the hand size remains the same, but the cards have changed
		utils.AssertEqual(t, newHandSize, oldHandSize)
		utils.AssertEqual(t, containsCard(newHand, targetCards...), false) // sometimes fails

		// And the pile has two extra cards (from the hand)
		utils.AssertTrue(t, newPileSize > oldPileSize)
		utils.AssertTrue(t, containsCard(game.pile, targetCards...))

		// And the deck has two fewer cards
		utils.AssertTrue(t, newDeckSize == oldDeckSize-2)
	})

	t.Run("stage 1: player picks up pile", func(t *testing.T) {
		// Given a game with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Clubs) // Ace of Clubs

		// and a player with low-value cards in their Hand
		lowValueCards := []deck.Card{
			deck.NewCard(deck.Four, deck.Hearts),
			deck.NewCard(deck.Five, deck.Clubs),
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		game := NewShed(ShedOpts{
			stage:           clearDeck,
			deck:            someDeck(4),
			pile:            []deck.Card{highValueCard},
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": &PlayerCards{Hand: deck.Deck(lowValueCards)},
				"player-2": somePlayerCards(3),
			},
		})

		oldHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		oldPileSize := len(game.pile)
		oldDeckSize := len(game.deck)

		// when a player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		newHandSize := len(game.playerCards[game.currentPlayerID].Hand)
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
		utils.AssertEqual(t, msgs[0].Command, protocol.SkipTurn)

		// and the other players' OutboundMessages have the expected content
		utils.AssertEqual(t, msgs[1].AwaitingResponse, false)
		utils.AssertEqual(t, msgs[1].Command, protocol.SkipTurn)

		// and the current player's response is handled correctly
		previousPlayerID := game.currentPlayerID
		response, err := game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.SkipTurn,
		}})
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, response, []OutboundMessage(nil))
		utils.AssertEqual(t, game.awaitingResponse, false)

		// and the next player is up
		utils.AssertTrue(t, game.currentPlayerID != previousPlayerID)
	})

	t.Run("stage 1: not enough cards in deck", func(t *testing.T) {
		// Given a game in stage 1 with one card left on the deck
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		targetCards := []deck.Card{
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Nine, deck.Diamonds),
		}

		// And a player with two cards of the same value in their hand
		pc := &PlayerCards{Hand: append(targetCards, deck.NewCard(deck.Eight, deck.Hearts))}

		game := NewShed(ShedOpts{
			stage:           clearDeck,
			deck:            someDeck(1),
			pile:            []deck.Card{lowValueCard},
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		oldHand := game.playerCards[game.currentPlayerID].Hand
		oldHandSize, oldPileSize := len(oldHand), len(game.pile)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		// And chooses to play two cards of the same value
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0, 1},
		}})
		utils.AssertNoError(t, err)

		newHand := game.playerCards[game.currentPlayerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.pile)
		newDeckSize := len(game.deck)

		// Then the hand size is smaller, and the cards have changed
		utils.AssertEqual(t, newHandSize, oldHandSize-1)
		utils.AssertEqual(t, containsCard(newHand, targetCards...), false)

		// And the pile has two extra cards (from the hand)
		utils.AssertEqual(t, newPileSize, oldPileSize+2)
		utils.AssertTrue(t, containsCard(game.pile, targetCards...))

		// And the deck is empty
		utils.AssertEqual(t, newDeckSize, 0)

		// And the game produces messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))
		checkReceiveResponseMessages(t, msgs, protocol.ReplenishHand, game)
		utils.AssertTrue(t, game.awaitingResponse)

		// And when the current player acks and releases their turn
		previousPlayerID := game.currentPlayerID
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.ReplenishHand,
		}})
		utils.AssertNoError(t, err)

		// And the game switches to stage 2
		utils.AssertEqual(t, game.stage, clearCards)

		// And the next player is up
		utils.AssertTrue(t, game.currentPlayerID != previousPlayerID)
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
			stage:           clearCards,
			deck:            deck.Deck{},
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		// When a player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertTrue(t, game.awaitingResponse)

		moves := getMoves(msgs, game.currentPlayerID)

		oldHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		oldSeenSize := len(game.playerCards[game.currentPlayerID].Seen)
		oldUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)
		previousPlayerID := game.currentPlayerID

		cardChoice := []int{moves[0]}
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, game.awaitingResponse)

		newHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		newSeenSize := len(game.playerCards[game.currentPlayerID].Seen)
		newUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)

		// Then their hand is smaller, but their remaining cards are unchanged
		utils.AssertTrue(t, newHandSize < oldHandSize)
		utils.AssertTrue(t, newSeenSize == oldSeenSize)
		utils.AssertTrue(t, newUnseenSize == oldUnseenSize)

		// And when the player releases their turn
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.EndOfTurn,
		}})
		utils.AssertNoError(t, err)

		// Then the game is no longer expecting a response
		utils.AssertEqual(t, game.awaitingResponse, false)

		// And it's the next player's turn
		utils.AssertTrue(t, previousPlayerID != game.currentPlayerID)
	})

	t.Run("stage 2: player has legal moves and no hand cards", func(t *testing.T) {
		// Given a game in stage 2, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Six, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		// And a player with an empty hand and a full set of Seen cards
		pc := &PlayerCards{
			Hand: []deck.Card{},
			Seen: []deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
		}
		game := NewShed(ShedOpts{
			stage:           clearCards,
			deck:            deck.Deck{},
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		// When the player starts their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlaySeen, game)

		moves := getMoves(msgs, game.currentPlayerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldSeenSize := len(game.playerCards[game.currentPlayerID].Seen)
		oldUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)
		oldPileSize := len(game.pile)

		// And when the player makes their choice
		cardChoice := []int{moves[0]}
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlaySeen,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		newHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		newSeenSize := len(game.playerCards[game.currentPlayerID].Seen)
		newUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)
		newPileSize := len(game.pile)

		// Then they have one less Seen card
		utils.AssertEqual(t, newSeenSize, oldSeenSize-1)
		// but their remaining cards are unchanged
		utils.AssertEqual(t, newHandSize, 0)
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize)
		// And the pile increases
		utils.AssertEqual(t, newPileSize, oldPileSize+1)

		// And everyone is informed
		for _, m := range msgs {
			utils.AssertEqual(t, m.Command, protocol.EndOfTurn)

			if m.PlayerID == game.currentPlayerID {
				utils.AssertTrue(t, m.AwaitingResponse)
			} else {
				utils.AssertEqual(t, m.AwaitingResponse, false)
			}
		}

		// And the game expects an ack from the player
		utils.AssertTrue(t, game.awaitingResponse)

		previousPlayerID := game.currentPlayerID
		// And when the player sends an ack
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.EndOfTurn,
		}})
		utils.AssertNoError(t, err)

		// Then it's the next player's turn
		utils.AssertEqual(t, game.awaitingResponse, false)
		utils.AssertTrue(t, previousPlayerID != game.currentPlayerID)
	})

	t.Run("stage 2: player has no legal moves and no hand cards", func(t *testing.T) {
		// Given a game in stage 2, with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Hearts)
		pile := []deck.Card{highValueCard}

		// And a player with an empty hand and a full set of seen cards
		pc := &PlayerCards{
			Hand: []deck.Card{},
			Seen: []deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
		}

		game := NewShed(ShedOpts{
			stage:           clearCards,
			deck:            deck.Deck{},
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})
		oldHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		oldPileSize := len(game.pile)
		oldSeenSize := len(game.playerCards[game.currentPlayerID].Seen)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		newHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		newPileSize := len(game.pile)
		newSeenSize := len(game.playerCards[game.currentPlayerID].Seen)

		// then the current player's hand includes the cards from the pile
		utils.AssertEqual(t, newHandSize, oldHandSize+oldPileSize)
		// and the pile is now empty
		utils.AssertEqual(t, newPileSize, 0)
		// and the seen cards are the same
		utils.AssertEqual(t, newSeenSize, oldSeenSize)

		// then everyone is informed
		utils.AssertTrue(t, len(msgs) > 0)
		utils.AssertEqual(t, len(msgs), len(game.playerIDs))

		// and the current player's OutboundMessage has the expected content
		utils.AssertTrue(t, msgs[0].AwaitingResponse)
		utils.AssertEqual(t, msgs[0].Command, protocol.SkipTurn)

		// and the other players' OutboundMessages have the expected content
		utils.AssertEqual(t, msgs[1].AwaitingResponse, false)
		utils.AssertEqual(t, msgs[1].Command, protocol.SkipTurn)

		// and the current player's response is handled correctly
		previousPlayerID := game.currentPlayerID
		response, err := game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.SkipTurn,
		}})
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, response, []OutboundMessage(nil))
		utils.AssertEqual(t, game.awaitingResponse, false)

		// and the next player is up
		utils.AssertTrue(t, game.currentPlayerID != previousPlayerID)
	})

	t.Run("stage 2: player only unseen cards", func(t *testing.T) {
		// Given a game in stage 2, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Six, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		// And a player with only a full set of Unseen cards
		pc := &PlayerCards{
			Hand: []deck.Card{},
			Seen: []deck.Card{},
			Unseen: []deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
		}

		game := NewShed(ShedOpts{
			stage:           clearCards,
			deck:            deck.Deck{},
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})
		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlayUnseen, game)

		// Then the game selects all Unseen cards (legal moves or not)
		moves := getMoves(msgs, game.currentPlayerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)
		previousPlayerID := game.currentPlayerID

		// And when the player selects a legal move
		cardChoice := []int{moves[0]}
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayUnseen,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.awaitingResponse, true)

		newHandSize := len(game.playerCards[game.currentPlayerID].Hand)
		newUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)

		// Then they have one less Unseen card, but their remaining cards are unchanged
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize-1)
		utils.AssertEqual(t, newHandSize, 0)

		// And everyone is informed of the end of turn
		checkReceiveResponseMessages(t, msgs, protocol.UnseenSuccess, game)

		// And the game expects an ack from the player
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.UnseenSuccess,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.awaitingResponse, false)

		// And it's the next player's turn
		utils.AssertTrue(t, previousPlayerID != game.currentPlayerID)
	})

	t.Run("stage 2: player has no legal moves and only unseen cards", func(t *testing.T) {
		// Given a game in stage 2
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{highValueCard}

		// And a player with only a full set of Unseen cards
		pc := &PlayerCards{
			Hand: []deck.Card{},
			Seen: []deck.Card{},
			Unseen: []deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
		}

		game := NewShed(ShedOpts{
			stage:           clearCards,
			deck:            deck.Deck{},
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
				"player-2": somePlayerCards(3),
			},
		})

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlayUnseen, game)

		// Then the game selects all Unseen cards (legal moves or not)
		moves := getMoves(msgs, game.currentPlayerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)
		oldPileSize := len(game.pile)

		// And when the player selects an illegal move
		cardChoice := []int{moves[0]}
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayUnseen,
			Decision: cardChoice,
		}})

		utils.AssertNoError(t, err)

		newHand := game.playerCards[game.currentPlayerID].Hand
		newUnseenSize := len(game.playerCards[game.currentPlayerID].Unseen)

		// Then the player picks up the pile and the Unseen cards are unaffected
		utils.AssertEqual(t, len(game.pile), 0)
		utils.AssertDeepEqual(t, len(newHand), oldPileSize)
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize)
		checkReceiveResponseMessages(t, msgs, protocol.UnseenFailure, game)

		// And the game is expecting an ack
		utils.AssertTrue(t, game.awaitingResponse)

		// And when the player's ack is received
		previousPlayerID := game.currentPlayerID
		msgs, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.UnseenFailure,
		}})

		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.awaitingResponse, false)

		// Then it's the next player's turn
		utils.AssertTrue(t, previousPlayerID != game.currentPlayerID)
	})

	t.Run("stage 2: player has legal moves clears their cards (finished game)", func(t *testing.T) {
		// Given a game in stage 2
		// and a player with one remaining Unseen card
		// When the player takes a legal turn
		// Then they have finished the game
		// And the game expects no response
		// And it's the next player's turn
	})

	t.Run("stage 2: game ends when n-1 players have finished", func(t *testing.T) {
		// Given a game in stage 2
		// and a player with one remaining Unseen card
		// When the player takes a legal turn
		// Then they have finished the game
		// And the game expects no response
		// And it's the next player's turn
	})
}

func TestGameBurn(t *testing.T) {
	// play a 10
	// pile has four of the same suit
	// play multiple 10s in one move == one burn
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

	// test contents of messages
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

			if playerIdx == game.currentTurnIdx {
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
			stage:           clearDeck,
			deck:            someDeck(1),
			pile:            pile,
			playerIDs:       []string{"player-1", "player-2"},
			currentPlayerID: "player-1",
			playerCards: map[string]*PlayerCards{
				"player-1": somePlayerCards(3),
				"player-2": somePlayerCards(3),
			},
		})

		// When the current player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)

		utils.AssertEqual(t, msgs[0].PlayerID, game.currentPlayerID)

		playerMoves := msgs[0].Moves
		utils.AssertTrue(t, len(playerMoves) > 0)

		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{playerMoves[0]}, // first possible move
		}})

		// Then the game expects an ack
		utils.AssertTrue(t, game.awaitingResponse)

		// And when the game receives the ack
		_, err = game.ReceiveResponse([]InboundMessage{{
			PlayerID: game.currentPlayerID,
			Command:  protocol.ReplenishHand,
		}})

		// Then the game stage switches to stage 2
		utils.AssertEqual(t, game.stage, clearCards)
	})
}

func checkNextMessages(t *testing.T, msgs []OutboundMessage, cmd protocol.Cmd, game *shed) {
	t.Helper()

	for _, m := range msgs {
		utils.AssertDeepEqual(t, m.Hand, game.playerCards[m.PlayerID].Hand)
		utils.AssertDeepEqual(t, m.Seen, game.playerCards[m.PlayerID].Seen)

		if m.PlayerID == game.currentPlayerID {
			// and the current player is asked to make a choice
			utils.AssertEqual(t, m.Command, cmd)
			utils.AssertTrue(t, m.AwaitingResponse)
			utils.AssertTrue(t, len(m.Moves) > 0)
		} else {
			cmdForOtherPlayers := protocol.Turn
			if cmd == protocol.ReplenishHand {
				cmdForOtherPlayers = protocol.EndOfTurn
			}
			utils.AssertEqual(t, m.Command, cmdForOtherPlayers)
			utils.AssertEqual(t, m.AwaitingResponse, false)
		}
	}
}

func checkReceiveResponseMessages(t *testing.T, msgs []OutboundMessage, currentPlayerCmd protocol.Cmd, game *shed) {
	t.Helper()

	for _, m := range msgs {
		utils.AssertDeepEqual(t, m.Hand, game.playerCards[m.PlayerID].Hand)
		utils.AssertDeepEqual(t, m.Seen, game.playerCards[m.PlayerID].Seen)

		if m.PlayerID == game.currentPlayerID {
			// and the current player is asked to make a choice
			utils.AssertEqual(t, m.Command, currentPlayerCmd)
			utils.AssertTrue(t, m.AwaitingResponse)
		} else {
			utils.AssertEqual(t, m.Command, protocol.EndOfTurn)
			utils.AssertEqual(t, m.AwaitingResponse, false)
		}
	}
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

func containsCard(s []deck.Card, targets ...deck.Card) bool {
	for _, c := range s {
		for _, tg := range targets {
			if c == tg {
				return true
			}
		}
	}
	return false
}

func getMoves(msgs []OutboundMessage, currentPlayerID string) []int {
	var moves []int
	for _, m := range msgs {
		if m.PlayerID == currentPlayerID {
			moves = m.Moves
		}
	}
	return moves
}
