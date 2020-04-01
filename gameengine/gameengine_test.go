package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/minaorangina/shed/player"
)

func TestGameEngine(t *testing.T) {
	// minimum of 2 players
	tooFewPlayers := []string{"Grace"}
	_, err := New(tooFewPlayers)
	if err == nil {
		t.Errorf("\nToo few players - expected GameEngine to return error, but it didn't")
	}

	// maximum of 4 players
	tooManyPlayers := []string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}
	_, err = New(tooManyPlayers)
	if err == nil {
		t.Errorf("\nToo many players - expected GameEngine to return error, but it didn't")
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
