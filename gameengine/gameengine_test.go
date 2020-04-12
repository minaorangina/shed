package gameengine

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/player"
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
	engine, err := New(playerNames)

	if err != nil {
		t.Fail()
	}
	expectedPlayers := []player.Player{{Name: "Ada"}, {Name: "Katherine"}}
	expectedEngine := GameEngine{
		players: expectedPlayers,
	}
	if !reflect.DeepEqual(expectedEngine, engine) {
		t.Errorf("\nExpected: GameEngine %+v\nActual: %+v", expectedEngine, engine)
	}

	// .Init
	// Should error if game not idle
	engineWithGameInProgress := GameEngine{
		players:   expectedPlayers,
		gameState: inProgress,
	}
	err = engineWithGameInProgress.Init()
	if err == nil {
		t.Errorf("\nExpected GameEngine.Init() to return an error (game not in idle state)")
	}

	engineWithGamePaused := GameEngine{
		players:   expectedPlayers,
		gameState: paused,
	}
	err = engineWithGamePaused.Init()
	if err == nil {
		t.Errorf("\nExpected GameEngine.Init() to return an error (game not in idle state)")
	}

	// Success case
	err = engine.Init()
	if err != nil {
		t.Errorf("\nFailed to call GameEngine.Init(): %s", err.Error())
	}
	fmt.Printf("after +%v\n", engine)
	expectedGameState := "inProgress"
	if engine.GameState() != expectedGameState {
		t.Errorf("\nExpected game state: %s\nActual: %s", expectedGameState, engine.GameState())
	}
}
