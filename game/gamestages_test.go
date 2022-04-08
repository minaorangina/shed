package game

import (
	"reflect"
	"testing"

	"github.com/minaorangina/shed/deck"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

func TestGameStageZero(t *testing.T) {
	t.Run("players asked to reorganise their hand", func(t *testing.T) {
		t.Log("When a new game starts")
		game, err := NewShed(threePlayers())
		utils.AssertNoError(t, err)

		msgs, err := game.Next()

		t.Log("Then players receive an instruction to reorganise their cards")
		utils.AssertNoError(t, err)
		utils.AssertTrue(t, len(msgs) == len(game.PlayerInfo))

		for _, m := range msgs {
			utils.AssertEqual(t, m.Command, protocol.Reorg)
		}

		t.Log("and the game is awaiting a response")
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Reorg)
	})

	t.Run("reorganised cards handled correctly", func(t *testing.T) {
		t.Log("Given a new game")
		game, err := NewShed(threePlayers())
		utils.AssertNoError(t, err)

		t.Log("When Next is called")
		_, err = game.Next()
		utils.AssertNoError(t, err)

		t.Log("Then the game enters the reorganisation stage")
		utils.AssertEqual(t, game.Stage, preGame)

		p2Cards := game.PlayerCards["p2"]
		p2NewCards := NewPlayerCards(
			[]deck.Card{p2Cards.Hand[2], p2Cards.Seen[1], p2Cards.Seen[2]},
			[]deck.Card{p2Cards.Hand[0], p2Cards.Hand[1], p2Cards.Seen[0]},
			nil, nil,
		)
		p2NewCards.Unseen = game.PlayerCards["p2"].Unseen
		p2NewCards.UnseenVisibility = game.PlayerCards["p2"].UnseenVisibility

		want := map[string]*PlayerCards{
			"p1": game.PlayerCards["p1"],
			"p2": p2NewCards,
			"p3": game.PlayerCards["p3"],
		}

		t.Log("and the players send their response")
		msgs, err := game.ReceiveResponse([]protocol.InboundMessage{
			{
				PlayerID: "p1",
				Command:  protocol.Reorg,
				Decision: []int{0, 1, 2},
			},
			{
				PlayerID: "p2",
				Command:  protocol.Reorg,
				Decision: []int{2, 4, 5},
			},
			{
				PlayerID: "p3",
				Command:  protocol.Reorg,
				Decision: []int{0, 1, 2},
			},
		})

		t.Log("and the response is accepted")
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, msgs, []protocol.OutboundMessage(nil))

		t.Log("and players' cards are updated")
		for id, c := range game.PlayerCards {
			utils.AssertDeepEqual(t, *c, *want[id])
		}

		t.Log("and the game moves to stage one")
		utils.AssertEqual(t, game.Stage, clearDeck)

		t.Log("and when Next() is called, someone is assigned the first turn")
		_, err = game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotEmptyString(t, game.CurrentPlayer.PlayerID)
	})
}

func TestGameStageZeroToOne(t *testing.T) {
	// Given a new game
	game := ExistingShed(ShedOpts{})

	// When the game has started and Next is called
	err := game.Start(twoPlayers())
	utils.AssertNoError(t, err)
	_, err = game.Next()
	utils.AssertNoError(t, err)

	// Then the game enters the reorganisation stage
	utils.AssertEqual(t, game.Stage, preGame)

	p2Cards := game.PlayerCards["p2"]
	p2NewCards := NewPlayerCards(
		[]deck.Card{p2Cards.Hand[2], p2Cards.Seen[1], p2Cards.Seen[2]},
		[]deck.Card{p2Cards.Hand[0], p2Cards.Hand[1], p2Cards.Seen[0]},
		nil, nil,
	)
	p2NewCards.Unseen = game.PlayerCards["p2"].Unseen
	p2NewCards.UnseenVisibility = game.PlayerCards["p2"].UnseenVisibility

	want := map[string]*PlayerCards{
		"p1": game.PlayerCards["p1"],
		"p2": p2NewCards,
	}

	// and the players send their response
	msgs, err := game.ReceiveResponse([]protocol.InboundMessage{
		{
			PlayerID: "p1",
			Command:  protocol.Reorg,
			Decision: []int{0, 1, 2},
		},
		{
			PlayerID: "p2",
			Command:  protocol.Reorg,
			Decision: []int{2, 4, 5},
		},
	})

	// and the response is accepted
	utils.AssertNoError(t, err)
	utils.AssertDeepEqual(t, msgs, []protocol.OutboundMessage(nil))

	// and players' cards are updated
	for id, c := range game.PlayerCards {
		utils.AssertDeepEqual(t, *c, *want[id])
	}

	// and the game moves to stage one
	utils.AssertEqual(t, game.Stage, clearDeck)

	// and when Next() is called, someone is assigned the first turn
	preMoveMsgs, err := game.Next()
	utils.AssertNoError(t, err)
	currentPlayerID := game.CurrentPlayer.PlayerID
	utils.AssertNotEmptyString(t, currentPlayerID)

	// And when the current player plays a card
	preMoveHand := make([]deck.Card, 3)
	copy(preMoveHand, game.PlayerCards[currentPlayerID].Hand)

	move := getNonBurnMove(game.PlayerCards[currentPlayerID].Hand, preMoveMsgs[0].Moves)
	chosenCard := preMoveHand[move]

	postMoveMsgs, err := game.ReceiveResponse([]protocol.InboundMessage{{
		PlayerID: game.CurrentPlayer.PlayerID,
		Command:  protocol.PlayHand,
		Decision: []int{move},
	}})

	// Then the current player's hand has changed
	utils.AssertNotDeepEqual(t, game.PlayerCards[currentPlayerID].Hand, preMoveHand)

	// And the messages reflect this fact
	utils.AssertNotDeepEqual(t, postMoveMsgs[0].Hand, preMoveHand)

	// And the pile is no longer empty
	utils.AssertEqual(t, len(game.Pile), 1)
	utils.AssertDeepEqual(t, game.Pile, []deck.Card{chosenCard})

	// And the messages reflect this fact
	for _, m := range postMoveMsgs {
		utils.AssertEqual(t, len(m.Pile), 1)
		utils.AssertDeepEqual(t, m.Pile, []deck.Card{chosenCard})
	}
}

func TestGameStageOne(t *testing.T) {
	t.Run("first move", func(t *testing.T) {
		// Given a game in stage 1 with an empty pile
		pile := []deck.Card{}

		// And a player with cards in their hand
		targetCard := deck.NewCard(deck.Nine, deck.Clubs)
		hand := []deck.Card{
			deck.NewCard(deck.Eight, deck.Hearts),
			targetCard,
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		pc := NewPlayerCards(hand, nil, nil, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(4),
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize, oldPileSize, oldDeckSize := len(oldHand), len(game.Pile), len(game.Deck)

		// When the game progresses, then players are informed of the current turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))

		checkNextMessages(t, msgs, protocol.PlayHand, game)
		moves := getMoves(msgs, game.CurrentPlayer.PlayerID)

		// And the game expects a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		// And when player response is received
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{moves[1]}, // target card is the second one
		}})
		utils.AssertNoError(t, err)

		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize, newPileSize, newDeckSize := len(newHand), len(game.Pile), len(game.Deck)

		// Then the pile contains the selected card
		utils.AssertTrue(t, newPileSize > oldPileSize)
		utils.AssertTrue(t, containsCard(game.Pile, targetCard))

		// And the deck decreases in size
		utils.AssertTrue(t, newDeckSize < oldDeckSize)

		// And the size of the player's hand remains the same
		utils.AssertTrue(t, newHandSize == oldHandSize)

		// But the cards in the player's hand changed
		utils.AssertEqual(t, reflect.DeepEqual(oldHand, newHand), false)

		// And all cards are unique
		utils.AssertTrue(t, cardsUnique(newHand)) // this fails sometimes
		utils.AssertTrue(t, cardsUnique(game.Pile))

		// And the game produces messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
		checkReceiveResponseMessages(t, msgs, protocol.ReplenishHand, game)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.ReplenishHand)

		// And when the current player acks
		previousPlayerID := game.CurrentPlayer.PlayerID
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.ReplenishHand,
		}})
		utils.AssertNoError(t, err)

		// Then their turn is released and the next player is up
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
	})

	t.Run("player has legal moves", func(t *testing.T) {
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

		pc := NewPlayerCards(hand, nil, nil, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(4),
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize, oldPileSize, oldDeckSize := len(oldHand), len(game.Pile), len(game.Deck)

		// When the game progresses, then players are informed of the current turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))

		checkNextMessages(t, msgs, protocol.PlayHand, game)
		moves := getMoves(msgs, game.CurrentPlayer.PlayerID)

		// And the game expects a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		// And when player response is received
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{moves[1]}, // target card is the second one
		}})
		utils.AssertNoError(t, err)

		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize, newPileSize, newDeckSize := len(newHand), len(game.Pile), len(game.Deck)

		// Then the pile contains the selected card
		utils.AssertTrue(t, newPileSize > oldPileSize)
		utils.AssertTrue(t, containsCard(game.Pile, targetCard))

		// And the deck decreases in size
		utils.AssertTrue(t, newDeckSize < oldDeckSize)

		// And the size of the player's hand remains the same
		utils.AssertTrue(t, newHandSize == oldHandSize)

		// But the cards in the player's hand changed
		utils.AssertEqual(t, reflect.DeepEqual(oldHand, newHand), false)

		// And all cards are unique
		utils.AssertTrue(t, cardsUnique(newHand)) // this fails sometimes
		utils.AssertTrue(t, cardsUnique(game.Pile))

		// And the game produces messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
		checkReceiveResponseMessages(t, msgs, protocol.ReplenishHand, game)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.ReplenishHand)

		// And when the current player acks and releases their turn
		previousPlayerID := game.CurrentPlayer.PlayerID
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.ReplenishHand,
		}})
		utils.AssertNoError(t, err)

		// And the next player is up
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
	})

	t.Run("player plays multiple cards of the same rank", func(t *testing.T) {
		// Given a game in stage 1
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}
		targetCards := []deck.Card{
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Nine, deck.Diamonds),
		}
		// And a player with two cards of the same value in their hand
		pc := NewPlayerCards(append(targetCards, deck.NewCard(deck.Eight, deck.Hearts)), nil, nil, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(4),
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize := len(oldHand)
		oldPileSize := len(game.Pile)
		oldDeckSize := len(game.Deck)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, game.CurrentTurnIdx, 0)
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != "")

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		// And chooses to play two of the same card
		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0, 1},
		}})
		utils.AssertNoError(t, err)
		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.Pile)
		newDeckSize := len(game.Deck)

		// Then the hand size remains the same, but the cards have changed
		utils.AssertEqual(t, newHandSize, oldHandSize)
		utils.AssertEqual(t, containsCard(newHand, targetCards...), false) // sometimes fails

		// And the pile has two extra cards (from the hand)
		utils.AssertTrue(t, newPileSize > oldPileSize)
		utils.AssertTrue(t, containsCard(game.Pile, targetCards...))

		// And the deck has two fewer cards
		utils.AssertTrue(t, newDeckSize == oldDeckSize-2)
	})

	t.Run("cannot play multiple cards from different ranks", func(t *testing.T) {
		// Given a game in stage 1
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}
		targetCards := []deck.Card{
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Six, deck.Clubs),
		}
		// And a player with cards different values in their hand
		pc := NewPlayerCards(
			append(targetCards, deck.NewCard(deck.Eight, deck.Hearts)),
			nil, nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(4),
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize := len(oldHand)
		oldPileSize := len(game.Pile)
		oldDeckSize := len(game.Deck)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, game.CurrentTurnIdx, 0)
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != "")

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		// And chooses to play two cards
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0, 1},
		}})

		// Then the game returns an error
		utils.AssertErrored(t, err)
		utils.AssertEqual(t, err, ErrInvalidMove)

		// And players are sent error messages
		utils.AssertTrue(t, len(msgs) > 0)
		for _, m := range msgs {
			utils.AssertEqual(t, m.Command, protocol.Error)
		}

		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.Pile)
		newDeckSize := len(game.Deck)

		// And the cards remain unchanged

		// And the hand size remains the same, but the cards have changed
		utils.AssertEqual(t, containsCard(game.Pile, targetCards...), false)
		utils.AssertEqual(t, newPileSize, oldPileSize)
		utils.AssertEqual(t, newDeckSize, oldDeckSize)
		utils.AssertEqual(t, newHandSize, oldHandSize)
		utils.AssertDeepEqual(t, newHand, oldHand)
	})

	t.Run("player picks up pile", func(t *testing.T) {
		// Given a game with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Clubs) // Ace of Clubs

		// and a player with low-value cards in their Hand
		lowValueCards := []deck.Card{
			deck.NewCard(deck.Four, deck.Hearts),
			deck.NewCard(deck.Five, deck.Clubs),
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(4),
			Pile:          []deck.Card{highValueCard},
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": {Hand: deck.Deck(lowValueCards)},
				"p2": somePlayerCards(3),
			},
		})

		oldHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		oldPileSize := len(game.Pile)
		oldDeckSize := len(game.Deck)

		// when a player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.SkipTurn)

		newHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		newPileSize := len(game.Pile)
		newDeckSize := len(game.Deck)

		// then the current player's hand includes the cards from the pile
		utils.AssertEqual(t, newHandSize, oldHandSize+oldPileSize)
		// and the pile is now empty
		utils.AssertEqual(t, newPileSize, 0)
		// and the deck is unchanged
		utils.AssertEqual(t, newDeckSize, oldDeckSize)

		// then everyone is informed
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))

		// and the current player's protocol.OutboundMessage has the expected content
		utils.AssertTrue(t, msgs[0].ShouldRespond)
		utils.AssertEqual(t, msgs[0].Command, protocol.SkipTurn)

		// and the other players' protocol.OutboundMessages have the expected content
		utils.AssertEqual(t, msgs[1].ShouldRespond, false)
		utils.AssertEqual(t, msgs[1].Command, protocol.SkipTurn)

		// and the current player's response is handled correctly
		previousPlayerID := game.CurrentPlayer.PlayerID
		response, err := game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.SkipTurn,
		}})
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, response, []protocol.OutboundMessage(nil))
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// and the next player is up
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
	})

	t.Run("not enough cards in deck", func(t *testing.T) {
		// Given a game in stage 1 with one card left on the deck
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		targetCards := []deck.Card{
			deck.NewCard(deck.Nine, deck.Clubs),
			deck.NewCard(deck.Nine, deck.Diamonds),
		}

		// And a player with two cards of the same value in their hand
		pc := NewPlayerCards(
			append(targetCards, deck.NewCard(deck.Eight, deck.Hearts)),
			nil, nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(1),
			Pile:          []deck.Card{lowValueCard},
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize, oldPileSize := len(oldHand), len(game.Pile)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		// And chooses to play two cards of the same value
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0, 1},
		}})
		utils.AssertNoError(t, err)

		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize := len(newHand)
		newPileSize := len(game.Pile)
		newDeckSize := len(game.Deck)

		// Then the hand size is smaller, and the cards have changed
		utils.AssertEqual(t, newHandSize, oldHandSize-1)
		utils.AssertEqual(t, containsCard(newHand, targetCards...), false) // fails sometimes

		// And the pile has two extra cards (from the hand)
		utils.AssertEqual(t, newPileSize, oldPileSize+2)
		utils.AssertTrue(t, containsCard(game.Pile, targetCards...))

		// And the deck is empty
		utils.AssertEqual(t, newDeckSize, 0)

		// And the game produces messages to all players
		// expecting a response only from the current player
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))
		checkReceiveResponseMessages(t, msgs, protocol.ReplenishHand, game)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.ReplenishHand)

		// And when the current player acks and releases their turn
		previousPlayerID := game.CurrentPlayer.PlayerID
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.ReplenishHand,
		}})
		utils.AssertNoError(t, err)

		// And the game switches to stage 2
		utils.AssertEqual(t, game.Stage, clearCards)

		// And the next player is up
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
	})

	t.Run("player has more than 3 hand cards", func(t *testing.T) {
		// Given a game in stage one
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)

		// and a player with 4 cards in their hand
		pc := NewPlayerCards(
			someCards(4),
			nil, nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearDeck,
			Deck:          someDeck(1),
			Pile:          []deck.Card{lowValueCard},
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)

		moves := msgs[0].Moves
		utils.AssertTrue(t, len(moves) > 1)

		oldHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		oldHandSize, oldDeckSize := len(oldHand), len(game.Deck)

		// And they play one card
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: []int{0},
		}})
		utils.AssertNoError(t, err)

		// Then they do NOT receive a new card
		newHand := game.PlayerCards[game.CurrentPlayer.PlayerID].Hand
		newHandSize := len(newHand)
		newDeckSize := len(game.Deck)

		utils.AssertEqual(t, newDeckSize, oldDeckSize)
		utils.AssertEqual(t, newHandSize, oldHandSize-1)
	})
}

// stage 1 to 2

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

		pc := NewPlayerCards(
			hand,
			nil, nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When a player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, msgs)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		moves := getMoves(msgs, game.CurrentPlayer.PlayerID)

		oldHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		oldSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)
		oldUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)
		previousPlayerID := game.CurrentPlayer.PlayerID

		cardChoice := []int{moves[0]}
		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.EndOfTurn)

		newHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		newSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)
		newUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)

		// Then their hand is smaller, but their remaining cards are unchanged
		utils.AssertTrue(t, newHandSize < oldHandSize)
		utils.AssertTrue(t, newSeenSize == oldSeenSize)
		utils.AssertTrue(t, newUnseenSize == oldUnseenSize)

		// And when the player releases their turn
		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.EndOfTurn,
		}})
		utils.AssertNoError(t, err)

		// Then the game is no longer expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// And it's the next player's turn
		utils.AssertTrue(t, previousPlayerID != game.CurrentPlayer.PlayerID)
	})

	t.Run("stage 2: player has legal moves and no hand cards", func(t *testing.T) {

		// Given a game in stage 2, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Six, deck.Hearts)
		pile := []deck.Card{lowValueCard}

		// And a player with an empty hand and a full set of Seen cards
		pc := NewPlayerCards(
			nil,
			[]deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
			someCards(3),
			nil,
		)
		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When the player starts their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlaySeen)

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlaySeen, game)

		moves := getMoves(msgs, game.CurrentPlayer.PlayerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)
		oldUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)
		oldPileSize := len(game.Pile)

		// And when the player makes their choice
		cardChoice := []int{moves[0]}
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlaySeen,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		newHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		newSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)
		newUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)
		newPileSize := len(game.Pile)

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

			if m.PlayerID == game.CurrentPlayer.PlayerID {
				utils.AssertTrue(t, m.ShouldRespond)
			} else {
				utils.AssertEqual(t, m.ShouldRespond, false)
			}
		}

		// And the game expects an ack from the player
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.EndOfTurn)

		previousPlayerID := game.CurrentPlayer.PlayerID
		// And when the player sends an ack
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.EndOfTurn,
		}})
		utils.AssertNoError(t, err)

		// Then it's the next player's turn
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)
		utils.AssertTrue(t, previousPlayerID != game.CurrentPlayer.PlayerID)
	})

	t.Run("stage 2: player has no legal moves and no hand cards", func(t *testing.T) {

		// Given a game in stage 2, with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Hearts)
		pile := []deck.Card{highValueCard}

		// And a player with an empty hand and a full set of seen cards
		pc := NewPlayerCards(
			nil,
			[]deck.Card{
				deck.NewCard(deck.Eight, deck.Hearts),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
			nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})
		oldHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		oldPileSize := len(game.Pile)
		oldSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		newHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		newPileSize := len(game.Pile)
		newSeenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Seen)

		// then the current player's hand includes the cards from the pile
		utils.AssertEqual(t, newHandSize, oldHandSize+oldPileSize)
		// and the pile is now empty
		utils.AssertEqual(t, newPileSize, 0)
		// and the seen cards are the same
		utils.AssertEqual(t, newSeenSize, oldSeenSize)

		// then everyone is informed
		utils.AssertTrue(t, len(msgs) > 0)
		utils.AssertEqual(t, len(msgs), len(game.PlayerInfo))

		// and the current player's protocol.OutboundMessage has the expected content
		utils.AssertTrue(t, msgs[0].ShouldRespond)
		utils.AssertEqual(t, msgs[0].Command, protocol.SkipTurn)

		// and the other players' protocol.OutboundMessages have the expected content
		utils.AssertEqual(t, msgs[1].ShouldRespond, false)
		utils.AssertEqual(t, msgs[1].Command, protocol.SkipTurn)

		// and the current player's response is handled correctly
		previousPlayerID := game.CurrentPlayer.PlayerID
		response, err := game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.SkipTurn,
		}})
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, response, []protocol.OutboundMessage(nil))
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// and the next player is up
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
	})

	t.Run("stage 2: player only unseen cards", func(t *testing.T) {

		// Given a game in stage 2, with a low-value card on the pile
		lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
		pile := []deck.Card{lowValueCard}
		chosenCard := deck.NewCard(deck.Eight, deck.Hearts)

		// And a player with only a full set of Unseen cards
		pc := NewPlayerCards(
			[]deck.Card{}, []deck.Card{},
			[]deck.Card{
				chosenCard,
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
			map[deck.Card]bool{
				chosenCard:                            false,
				deck.NewCard(deck.Nine, deck.Clubs):   false,
				deck.NewCard(deck.Six, deck.Diamonds): false,
			},
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})
		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlayUnseen, game)

		// Then the game selects all Unseen cards (legal moves or not)
		moves := getMoves(msgs, game.CurrentPlayer.PlayerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)

		// And when the player selects a legal move
		decision := []int{moves[0]}

		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayUnseen,
			Decision: decision,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.UnseenSuccess)

		newHandSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Hand)
		newUnseenSize := len(game.PlayerCards[game.CurrentPlayer.PlayerID].Unseen)

		// Then they have the same number of Unseen cards
		// but their chosen card is now visible
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize)
		utils.AssertEqual(t, newHandSize, 0)
		for card, visible := range game.PlayerCards["p1"].UnseenVisibility {
			if card == chosenCard {
				utils.AssertTrue(t, visible)
			} else {
				utils.AssertEqual(t, visible, false)
			}
		}

		// And everyone is informed of the successful move
		checkReceiveResponseMessages(t, msgs, protocol.UnseenSuccess, game)

		playerID := game.CurrentPlayer.PlayerID

		// And the game expects an ack from the player
		_, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: playerID,
			Command:  protocol.UnseenSuccess,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		newHandSize = len(game.PlayerCards[playerID].Hand)
		newUnseenSize = len(game.PlayerCards[playerID].Unseen)
		// Then they have one less Unseen card, but their remaining cards are unchanged
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize-1)
		utils.AssertEqual(t, newHandSize, 0)
		// And it's the next player's turn
		utils.AssertTrue(t, playerID != game.CurrentPlayer.PlayerID)
	})

	t.Run("stage 2: player has no legal moves and only unseen cards", func(t *testing.T) {
		// Given a game in stage 2
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{highValueCard}
		chosenCard := deck.NewCard(deck.Eight, deck.Hearts)

		// And a player with only a full set of Unseen cards
		pc := NewPlayerCards(
			[]deck.Card{},
			[]deck.Card{},
			[]deck.Card{
				chosenCard,
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Six, deck.Diamonds),
			},
			map[deck.Card]bool{
				chosenCard:                            false,
				deck.NewCard(deck.Nine, deck.Clubs):   false,
				deck.NewCard(deck.Six, deck.Diamonds): false,
			},
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When the player takes their turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		playerID := game.CurrentPlayer.PlayerID

		// Then everyone is informed
		checkNextMessages(t, msgs, protocol.PlayUnseen, game)

		// Then the game selects all Unseen cards (legal moves or not)
		moves := getMoves(msgs, playerID)
		utils.AssertTrue(t, len(moves) > 0)
		utils.AssertDeepEqual(t, moves, []int{0, 1, 2})

		oldUnseenSize := len(game.PlayerCards[playerID].Unseen)
		oldPileSize := len(game.Pile)
		oldHand := game.PlayerCards[playerID].Hand

		// And when the player selects an illegal move
		decision := moves[0:1]
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: playerID,
			Command:  protocol.PlayUnseen,
			Decision: decision,
		}})

		utils.AssertNoError(t, err)
		checkReceiveResponseMessages(t, msgs, protocol.UnseenFailure, game)

		newUnseenSize := len(game.PlayerCards[playerID].Unseen)
		newPileSize := len(game.Pile)
		newHand := game.PlayerCards[playerID].Hand

		// Then none of the cards have changed
		utils.AssertEqual(t, oldUnseenSize, newUnseenSize)
		utils.AssertEqual(t, oldPileSize, newPileSize)
		utils.AssertDeepEqual(t, oldHand, newHand)
		// And the player's chosen card has been flipped
		utils.AssertTrue(t, game.PlayerCards[playerID].UnseenVisibility[chosenCard])

		// check which messages are sent out

		// And the game is expecting an ack
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.UnseenFailure)

		// And when the player's ack is received
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.UnseenFailure,
		}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, len(msgs), 0)

		newUnseenSize = len(game.PlayerCards[playerID].Unseen)
		newPileSize = len(game.Pile)
		newHand = game.PlayerCards[playerID].Hand

		// Then the player picks up the pile which includes the chosen Unseen card
		utils.AssertEqual(t, len(game.Pile), 0)
		utils.AssertDeepEqual(t, len(newHand), oldPileSize+1)
		utils.AssertEqual(t, newUnseenSize, oldUnseenSize-1)

		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// Then it's the next player's turn
		utils.AssertTrue(t, playerID != game.CurrentPlayer.PlayerID)
	})

	t.Run("stage 2: player finishes with final Unseen card", func(t *testing.T) {
		// Given a game in stage 2
		lowValueCard := deck.NewCard(deck.Four, deck.Spades)
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{lowValueCard}

		// And a player with one remaining Unseen card
		pc := NewPlayerCards(nil, nil, []deck.Card{highValueCard}, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       threePlayers(),
			CurrentPlayer: threePlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
				"p3": somePlayerCards(3),
			},
		})

		// When the player takes a legal turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		playerID := game.CurrentPlayer.PlayerID
		oldNumActivePlayers := len(game.ActivePlayers)

		cardChoice := []int{0}
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: playerID,
			Command:  protocol.PlayUnseen,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		// Then the players are informed of the successful move
		checkReceiveResponseMessages(t, msgs, protocol.UnseenSuccess, game)
		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.UnseenSuccess)

		// And when the player acks
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.UnseenSuccess,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the player has finished
		checkPlayerFinishedMessages(t, msgs, game)

		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayerFinished)

		// And when the player acks again
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayerFinished,
		}})
		utils.AssertNoError(t, err)
		newNumActivePlayers := len(game.ActivePlayers)

		// Then there is one less active players
		utils.AssertEqual(t, newNumActivePlayers, oldNumActivePlayers-1)
		utils.AssertTrue(t, sliceContainsPlayerID(game.FinishedPlayers, playerID))

		// And it's the next player's turn
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != playerID)
	})

	t.Run("stage 2: player finishes with final Hand card", func(t *testing.T) {

		// Given a game in stage 2
		lowValueCard := deck.NewCard(deck.Four, deck.Spades)
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{lowValueCard}

		// And a player with one remaining Hand card and no Unseen cards
		pc := NewPlayerCards(
			[]deck.Card{highValueCard},
			nil, nil, nil,
		)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       threePlayers(),
			CurrentPlayer: threePlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
				"p3": somePlayerCards(3),
			},
		})

		// When the player takes a legal turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		previousPlayerID := game.CurrentPlayer.PlayerID
		previousNumPlayers := len(game.ActivePlayers)

		cardChoice := []int{0}
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the player has finished
		checkPlayerFinishedMessages(t, msgs, game)

		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayerFinished)

		// And when the player acks
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayerFinished,
		}})
		utils.AssertNoError(t, err)

		// Then there is one less active players
		utils.AssertEqual(t, len(game.ActivePlayers), previousNumPlayers-1)
		utils.AssertTrue(t, sliceContainsPlayerID(game.FinishedPlayers, previousPlayerID))

		// And it's the next player's turn
		utils.AssertTrue(t, game.CurrentPlayer.PlayerID != previousPlayerID)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)
	})
}

// this may not belong here
func TestGameStageTwoToGameOver(t *testing.T) {
	t.Run("stage 2: game ends when n-1 players have finished (Unseen card)", func(t *testing.T) {
		// Given a game in stage 2
		lowValueCard := deck.NewCard(deck.Four, deck.Spades)
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{lowValueCard}

		// And a player with one remaining Unseen card
		pc := NewPlayerCards(nil, nil, []deck.Card{highValueCard}, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When the player takes a legal turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)

		cardChoice := []int{0}
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayUnseen,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		// Then the players are informed of the successful move
		checkReceiveResponseMessages(t, msgs, protocol.UnseenSuccess, game)
		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.UnseenSuccess)

		// And when the player acks
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.UnseenSuccess,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the player has finished
		checkPlayerFinishedMessages(t, msgs, game)

		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayerFinished)

		// And when the player acks again
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayerFinished,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the game is over
		checkGameOverMessages(t, msgs, game)

		// And the game is NOT expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// And calling game.Next() returns the same game over message
		msgs, err = game.Next()
		utils.AssertNoError(t, err)
		checkGameOverMessages(t, msgs, game)
	})

	t.Run("stage 2: game ends when n-1 players have finished (Hand card)", func(t *testing.T) {

		// Given a game in stage 2 with two players remaining
		lowValueCard := deck.NewCard(deck.Four, deck.Spades)
		highValueCard := deck.NewCard(deck.Ace, deck.Spades)
		pile := []deck.Card{lowValueCard}

		// And a player with one remaining Hand card
		pc := NewPlayerCards([]deck.Card{highValueCard}, nil, nil, nil)

		game := ExistingShed(ShedOpts{
			Stage:         clearCards,
			Deck:          deck.Deck{},
			Pile:          pile,
			Players:       twoPlayers(),
			CurrentPlayer: twoPlayers()[0],
			PlayerCards: map[string]*PlayerCards{
				"p1": pc,
				"p2": somePlayerCards(3),
			},
		})

		// When the player takes a legal turn
		msgs, err := game.Next()
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayHand)

		cardChoice := []int{0}
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayHand,
			Decision: cardChoice,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the player has finished
		checkPlayerFinishedMessages(t, msgs, game)

		// And the game is expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.PlayerFinished)

		// And when the player acks
		msgs, err = game.ReceiveResponse([]protocol.InboundMessage{{
			PlayerID: game.CurrentPlayer.PlayerID,
			Command:  protocol.PlayerFinished,
		}})
		utils.AssertNoError(t, err)

		// Then the game informs everyone the game is over
		checkGameOverMessages(t, msgs, game)

		// And the game is NOT expecting a response
		utils.AssertEqual(t, game.AwaitingResponse(), protocol.Null)

		// And calling game.Next() returns the same game over message
		msgs, err = game.Next()
		utils.AssertNoError(t, err)
		checkGameOverMessages(t, msgs, game)
	})
}
