package gameengine

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestGameEngine(t *testing.T) {
	type engineTest struct {
		testName string
		input    []string
		expected error
	}

	testsShouldError := []engineTest{
		{
			"too few players",
			[]string{"Grace"},
			errors.New("Could not construct GameEngine: minimum of 2 players required (supplied 1)"),
		},
		{
			"too many players",
			[]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"},
			errors.New("Could not construct GameEngine: maximum of 4 players required (supplied 5)"),
		},
	}

	for _, et := range testsShouldError {
		_, err := New(et.input)
		if err == nil {
			t.Errorf(utils.TableFailureMessage(et.testName, strings.Join(et.input, ","), et.expected.Error()))
		}
	}

	// Construct a GameEngine
	playerNames := []string{"Ada", "Katherine"}
	_, err := New(playerNames)
	if err != nil {
		t.Fail()
	}
}

func TestGameEngineInit(t *testing.T) {
	playerNames := []string{"Ada", "Katherine"}
	engine, err := New(playerNames)
	if err != nil {
		t.Fail()
	}
	// Init works
	engine.Init()
	if engine.game == nil {
		t.Errorf("engine.game was nil")
	}
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// Game Engine receives from messages to send to players
	// and returns response
	playerNames := []string{"Ada", "Katherine"}
	engine, err := New(playerNames)
	// Init works
	engine.Init()

	game := *engine.game
	game.start() // deal cards

	messages := make(map[string]messageToPlayer)
	expected := map[string]reorganisedHand{}
	for _, p := range *game.players {
		o := buildOpponents(p.id, *game.players)
		m := game.buildMessageToPlayer(p, o, "Rearrange your hand")
		messages[p.id] = m

		expected[p.id] = reorganisedHand{
			PlayerID:  p.id,
			HandCards: p.hand,
			SeenCards: p.seen,
		}
	}

	actual, err := engine.messagePlayersAwaitReply(messages)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(utils.FailureMessage(expected, actual))
	}
}
