package gameengine

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func gameEngineWithPlayers() (*GameEngine, AllPlayers) {
	player1 := NewPlayer(NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := NewPlayer(NewID(), "Sally", os.Stdin, os.Stdout)
	players := NewAllPlayers(player1, player2)

	ge, _ := New(players)
	return ge, players
}

// DO NOT RUN
func TestNewGameEngine(t *testing.T) {
	t.Skip("do not run TestNewGameEngine")
	type gameTest struct {
		testName string
		input    AllPlayers
		expected error
	}

	testsShouldError := []gameTest{
		{
			"too few players",
			namesToAllPlayers([]string{"Grace"}),
			errors.New("Could not construct Game: minimum of 2 players required (supplied 1)"),
		},
		{
			"too many players",
			namesToAllPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
			errors.New("Could not construct Game: maximum of 4 players required (supplied 5)"),
		},
	}

	for _, et := range testsShouldError {
		_, err := New(et.input)
		if err == nil {
			t.Errorf(utils.TableFailureMessage(et.testName, strings.Join(allPlayersToNames(et.input), ","), et.expected.Error()))
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
		c := p.cards()
		numHand := len(c.hand)
		numSeen := len(c.seen)
		numUnseen := len(c.unseen)
		if numHand != 3 {
			formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
			t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
		}
	}
}

func TestBuildMessageToPlayer(t *testing.T) {
	ge, players := gameEngineWithPlayers()
	var opponents []opponent
	var id string
	for key := range players {
		opponents = buildOpponents(key, players)
		id = key
		break
	}

	playerToContact := ge.players[id]
	message := ge.buildReorgMessage(playerToContact, opponents, initialCards{}, "Let the games begin!")
	expectedMessage := messageToPlayer{
		Message:   "Let the games begin!",
		PlayState: ge.playState,
		GameStage: ge.stage,
		PlayerID:  playerToContact.ID,
		Name:      playerToContact.Name,
		Hand:      playerToContact.cards().hand,
		Seen:      playerToContact.cards().seen,
		Opponents: opponents,
		Command:   reorg,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		t.Errorf(utils.FailureMessage(expectedMessage, message))
	}
}
