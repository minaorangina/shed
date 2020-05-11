package gameengine

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func gameWithPlayers() (*Game, map[string]*Player) {
	gameEngine := New()
	player1 := NewPlayer(NewID(), "Harry")
	player2 := NewPlayer(NewID(), "Sally")
	players := map[string]*Player{player1.id: &player1, player2.id: &player2}
	game, _ := NewGame(&gameEngine, []playerInfo{{player1.id, player1.name}, {player2.id, player2.name}})
	return game, players
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

	game, _ := gameWithPlayers()
	if len(game.deck) != 52 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 52, len(game.deck)))
	}
	if len(game.players) != 2 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 2, len(game.players)))
	}

	expectedStage := "handOrganisation"
	if game.Stage() != expectedStage {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", expectedStage, game.Stage()))
	}
}
func TestGameStart(t *testing.T) {
	game, _ := gameWithPlayers()

	game.start()

	for _, p := range game.players {
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
	game, players := gameWithPlayers()
	var opponents []opponent
	var id string
	for key := range players {
		opponents = buildOpponents(key, players)
		id = key
		break
	}

	playerToContact := game.players[id]
	message := game.buildMessageToPlayer(playerToContact, opponents, "Let the games begin!")
	expectedMessage := messageToPlayer{
		Message:   "Let the games begin!",
		PlayState: game.engine.playState,
		GameStage: game.stage,
		PlayerID:  playerToContact.id,
		Name:      playerToContact.name,
		HandCards: playerToContact.cards().hand,
		SeenCards: playerToContact.cards().seen,
		Opponents: opponents,
		Command:   reorg,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		t.Errorf(utils.FailureMessage(expectedMessage, message))
	}
}
