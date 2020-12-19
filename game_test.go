package shed

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/minaorangina/shed/deck"
	utils "github.com/minaorangina/shed/internal"
)

func TestShedStart(t *testing.T) {
	t.Run("only starts with legal number of players", func(t *testing.T) {

		tt := []struct {
			name      string
			playerIDs []string
			err       error
		}{
			{
				"too few players",
				[]string{"Grace"},
				ErrTooFewPlayers,
			},
			{
				"too many players",
				[]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"},
				ErrTooManyPlayers,
			},
			{
				"just right",
				[]string{"Ada", "Katherine", "Grace", "Hedy"},
				nil,
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				game := NewShed()
				err := game.Start(tc.playerIDs)

				utils.AssertDeepEqual(t, err, tc.err)
			})
		}
	})
}

func TestBuildMessageToPlayer(t *testing.T) {
	ge := gameEngineWithPlayers()
	ps := ge.Players()
	var opponents []Opponent
	var id string
	for _, p := range ps {
		id = p.ID()
		opponents = buildOpponents(id, ps)
		break
	}

	playerToContact, _ := ps.Find(id)
	message := buildReorgMessage(playerToContact, opponents, InitialCards{})
	expectedMessage := OutboundMessage{
		PlayerID:  playerToContact.ID(),
		Name:      playerToContact.Name(),
		Hand:      playerToContact.Cards().Hand,
		Seen:      playerToContact.Cards().Seen,
		Opponents: opponents,
		Command:   protocol.Reorg,
	}
	utils.AssertDeepEqual(t, message, expectedMessage)
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// update when GameEngine forwards messages from players

	t.Skip("do not run TestGameEngineMsgFromGame")
	ge := gameEngineWithPlayers()
	ge.Start() // mock required

	messages := []OutboundMessage{}
	want := []InboundMessage{}
	initialCards := InitialCards{}
	for _, p := range ge.Players() {
		o := buildOpponents(p.ID(), ge.Players())
		m := buildReorgMessage(p, o, initialCards)
		messages = append(messages, m)

		cards := p.Cards()

		want = append(want, InboundMessage{
			PlayerID: p.ID(),
			Hand:     cards.Hand,
			Seen:     cards.Seen,
		})
	}

	err := ge.MessagePlayers(messages)
	if err != nil {
		t.Fail()
	}
}

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

func TestGameStageOneLegalMoves(t *testing.T) {
	// Given a game with a low-value card on the pile
	lowValueCard := deck.NewCard(deck.Four, deck.Hearts)
	pile := []deck.Card{lowValueCard}

	// and a player with higher-value cards in their hand
	targetCard := deck.NewCard(deck.Nine, deck.Clubs)
	hand := []deck.Card{
		deck.NewCard(deck.Eight, deck.Hearts),
		targetCard,
		deck.NewCard(deck.Six, deck.Diamonds),
	}

	pc := &PlayerCards{Hand: deck.Deck(hand)}

	game := NewShed(ShedOpts{
		stage:     clearDeck,
		deck:      someDeck(4),
		pile:      pile,
		playerIDs: []string{"player-1", "player-2"},
		playerCards: map[string]*PlayerCards{
			"player-1": pc,
		},
	})

	playerID := "player-1"
	oldHand := game.playerCards[playerID].Hand
	oldHandSize := len(oldHand)
	oldPileSize := len(game.pile)
	oldDeckSize := len(game.deck)

	// when a player takes their turn
	msg, err := game.Next()
	utils.AssertNoError(t, err)

	// then the player is asked to make a choice
	utils.AssertNotNil(t, msg)
	utils.AssertTrue(t, msg[0].ExpectResponse)

	// and the game expects a response
	utils.AssertTrue(t, game.awaitingResponse)

	// and when player response is received
	msgs, err := game.ReceiveResponse([]InboundMessage{{
		PlayerID: playerID,
		Command:  protocol.PlayHand,
		Decision: []int{1},
	}})
	utils.AssertNoError(t, err)

	newHand := game.playerCards[playerID].Hand
	newHandSize := len(newHand)
	newPileSize := len(game.pile)
	newDeckSize := len(game.deck)

	// then the pile contains the selected card
	utils.AssertTrue(t, newPileSize > oldPileSize)
	utils.AssertTrue(t, containsCard(game.pile, targetCard))

	// and the deck decreases in size
	utils.AssertTrue(t, newDeckSize < oldDeckSize)

	// and the size of the player's hand remains the same
	utils.AssertTrue(t, newHandSize == oldHandSize)

	// but the cards in the player's hand changed
	utils.AssertEqual(t, reflect.DeepEqual(oldHand, newHand), false)

	// and all cards are unique
	utils.AssertTrue(t, cardsUnique(newHand))
	utils.AssertTrue(t, cardsUnique(game.pile))

	// and the game produces messages to all players, expecting no response
	utils.AssertNotNil(t, msgs)
	utils.AssertEqual(t, len(msgs), len(game.playerIDs))
	for _, m := range msgs {
		utils.AssertEqual(t, m.ExpectResponse, false)
	}
	utils.AssertEqual(t, game.awaitingResponse, false)
}
func TestGameNoLegalMoves(t *testing.T) {
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
			stage:     clearDeck,
			deck:      someDeck(4),
			pile:      []deck.Card{highValueCard},
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": &PlayerCards{Hand: deck.Deck(lowValueCards)},
			},
		})

		playerID := "player-1"
		oldHandSize := len(game.playerCards[playerID].Hand)
		oldPileSize := len(game.pile)
		oldDeckSize := len(game.deck)

		// when a player takes their turn
		msg, err := game.Next()
		utils.AssertNoError(t, err)
		_ = msg

		newHandSize := len(game.playerCards[playerID].Hand)
		newPileSize := len(game.pile)
		newDeckSize := len(game.deck)

		// then the hand includes the cards from the pile
		utils.AssertEqual(t, newHandSize, oldHandSize+oldPileSize)
		// and the pile is now empty
		utils.AssertEqual(t, newPileSize, 0)
		// and the deck is unchanged
		utils.AssertEqual(t, newDeckSize, oldDeckSize)

		// and the []OutboundMessage does not require a response
		utils.AssertNotNil(t, msg)
		utils.AssertEqual(t, msg[0].ExpectResponse, false)
	})
}

func TestGameInput(t *testing.T) {
	// test that the game won't proceed if it's expecting a response

	// test the game won't allow .Next() to be called if someone's turn is incomplete
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
				utils.AssertDeepEqual(t, legalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})
}

func someDeck(num int) deck.Deck {
	d := deck.New()
	return deck.Deck(d.Deal(num))
}

func containsCard(s []deck.Card, target deck.Card) bool {
	for _, c := range s {
		if c == target {
			return true
		}
	}
	return false
}
