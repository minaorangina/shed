package gameengine

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/protocol"
)

func TestNewGameEngine(t *testing.T) {
	t.Run("constructed with correct number of players", func(t *testing.T) {
		type gameTest struct {
			testName string
			input    players.Players
			expected error
		}
		testsShouldError := []gameTest{
			{
				"too few players",
				namesToPlayers([]string{"Grace"}),
				errors.New("Could not construct Game: minimum of 2 players required (supplied 1)"),
			},
			{
				"too many players",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
				errors.New("Could not construct Game: maximum of 4 players required (supplied 5)"),
			},
		}

		for _, et := range testsShouldError {
			_, err := New(et.input, nil)
			if err == nil {
				utils.TableFailureMessage(
					t,
					et.testName,
					strings.Join(playersToNames(et.input), ","),
					et.expected.Error(),
				)
			}
		}
	})

	t.Run("constructs with correct number of cards", func(t *testing.T) {
		ge, _ := gameEngineWithPlayers()
		if len(ge.Deck()) != 52 {
			t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 52, len(ge.Deck())))
		}
	})

	t.Run("unnamed for now", func(t *testing.T) {
		t.Skip("do not run TestGameStart")
		ge, players := gameEngineWithPlayers()

		err := ge.Start() // mock required
		if err != nil {
			t.Fatalf("Could not start game")
		}

		for _, p := range players {
			c := p.Cards()
			numHand := len(c.Hand)
			numSeen := len(c.Seen)
			numUnseen := len(c.Unseen)
			if numHand != 3 {
				formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
				t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
			}
		}
	})
}

func TestBuildMessageToPlayer(t *testing.T) {
	ge, ps := gameEngineWithPlayers()
	var opponents []players.Opponent
	var id string
	for _, p := range ps {
		id = p.ID
		opponents = buildOpponents(id, ps)
		break
	}

	playerToContact, _ := ge.Players().Find(id)
	message := buildReorgMessage(playerToContact, opponents, players.InitialCards{}, "Let the games begin!")
	expectedMessage := players.OutboundMessage{
		Message:   "Let the games begin!",
		PlayerID:  playerToContact.ID,
		Name:      playerToContact.Name,
		Hand:      playerToContact.Cards().Hand,
		Seen:      playerToContact.Cards().Seen,
		Opponents: opponents,
		Command:   protocol.Reorg,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		utils.FailureMessage(t, expectedMessage, message)
	}
}
