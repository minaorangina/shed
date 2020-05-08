package gameengine

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func gameWithPlayers() *Game {
	gameEngine := New()
	game, _ := NewGame(&gameEngine, []playerInfo{{"Harry-1", "Harry"}, {"Sally-1", "Sally"}})
	return game
}

func TestNewGame(t *testing.T) {
	type gameTest struct {
		testName string
		input    []string
		expected error
	}

	ge := New()

	testsShouldError := []gameTest{
		{
			"too few players",
			[]string{"Grace"},
			errors.New("Could not construct Game: minimum of 2 players required (supplied 1)"),
		},
		{
			"too many players",
			[]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"},
			errors.New("Could not construct Game: maximum of 4 players required (supplied 5)"),
		},
	}

	for _, et := range testsShouldError {
		err := ge.Init(et.input)
		if err == nil {
			t.Errorf(utils.TableFailureMessage(et.testName, strings.Join(et.input, ","), et.expected.Error()))
		}
	}

	game := gameWithPlayers()
	if len(game.deck) != 52 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 52, len(game.deck)))
	}
	if len(*game.players) != 2 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 2, len(*game.players)))
	}

	expectedStage := "handOrganisation"
	if game.Stage() != expectedStage {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", expectedStage, game.Stage()))
	}
}
func TestGameStart(t *testing.T) {
	game := gameWithPlayers()

	game.start()

	for _, p := range *game.players {
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
	game := gameWithPlayers()
	player0 := NewPlayer("hermy-0", "Hermione")
	player2 := NewPlayer("sally-1", "Sally")
	someMorePlayers := []Player{player0, player2}

	expectedOpponents := []opponent{
		{ID: player0.id, Name: player0.name, SeenCards: player0.cards().seen},
		{ID: player2.id, Name: player2.name, SeenCards: player2.cards().seen},
	}
	opponents := buildOpponents("harry-1", someMorePlayers)
	if !reflect.DeepEqual(opponents, expectedOpponents) {
		t.Errorf("\nExpected: %+v\nActual: %+v\n", expectedOpponents, opponents)
	}

	playerToContact := (*game.players)[1]
	message := game.buildMessageToPlayer(playerToContact, opponents, "Let the games begin!")
	expectedMessage := messageToPlayer{
		Message:   "Let the games begin!",
		PlayState: game.engine.playState,
		GameStage: game.stage,
		PlayerID:  playerToContact.id,
		HandCards: playerToContact.cards().hand,
		SeenCards: playerToContact.cards().seen,
		Opponents: expectedOpponents,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		t.Errorf(utils.FailureMessage(expectedMessage, message))
	}
}
