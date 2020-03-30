package gameengine

import (
	"reflect"
	"testing"

	"github.com/minaorangina/shed/player"
)

func TestGameEngine(t *testing.T) {
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
}
