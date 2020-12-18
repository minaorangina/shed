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
		t.Skip()
		// Given a game with a high-value card on the pile
		highValueCard := deck.NewCard(deck.Ace, deck.Clubs) // Ace of Clubs

		// and a player with low-value cards in their Hand
		lowValueCards := []deck.Card{
			deck.NewCard(deck.Four, deck.Hearts),
			deck.NewCard(deck.Five, deck.Clubs),
			deck.NewCard(deck.Six, deck.Diamonds),
		}

		pc := &PlayerCards{Hand: deck.Deck(lowValueCards)}

		game := NewShed(ShedOpts{
			stage:     clearDeck,
			deck:      someDeck(4),
			pile:      []deck.Card{highValueCard},
			playerIDs: []string{"player-1", "player-2"},
			playerCards: map[string]*PlayerCards{
				"player-1": pc,
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
	t.Run("fours", func(t *testing.T) {
		tt := []struct {
			name         string
			pile, toPlay []deck.Card
			moves        []int
		}{
			{
				name:   "four of ♠ vs five of ♣",
				pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				moves:  []int{},
			},
			{
				name:   "four of ♥ vs King of ♣",
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{},
			},
			{
				name:   "four of ♦ vs Seven of ♥",
				pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{},
			},
			{
				name:   "four of ♣ vs Ace of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{},
			},
			{
				name:   "four of ♣ vs Two of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{},
			},
			{
				name:   "four of ♣ vs four of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "four of ♥ vs four of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ vs four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ vs four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, legalMoves(tc.toPlay, tc.pile), tc.moves)
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
