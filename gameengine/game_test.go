package gameengine

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/protocol"
)

func gameEngineWithPlayers() (*GameEngine, players.Players) {
	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)
	players := players.Players([]*players.Player{player1, player2})

	ge, _ := New(players)
	return ge, players
}

// DO NOT RUN
func TestNewGameEngine(t *testing.T) {
	t.Skip("do not run TestNewGameEngine")
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
		_, err := New(et.input)
		if err == nil {
			t.Errorf(utils.TableFailureMessage(
				et.testName,
				strings.Join(playersToNames(et.input), ","),
				et.expected.Error()),
			)
		}
	}

	ge, players := gameEngineWithPlayers()
	if len(ge.deck) != 52 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 52, len(ge.deck)))
	}
	if len(players) != 2 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 2, len(ge.players)))
	}

	expectedStage := "cardOrganisation"
	if ge.stage.String() != expectedStage {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", expectedStage, ge.stage.String()))
	}
}
func TestGameStart(t *testing.T) {
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

	playerToContact, _ := ge.players.Individual(id)
	message := ge.buildReorgMessage(playerToContact, opponents, players.InitialCards{}, "Let the games begin!")
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
		t.Errorf(utils.FailureMessage(expectedMessage, message))
	}
}
