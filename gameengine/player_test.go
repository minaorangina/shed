package gameengine

import (
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p := NewPlayer("player-1", name)
	expectedPlayer := Player{
		name: name,
		id:   "player-1",
	}
	if !reflect.DeepEqual(expectedPlayer, p) {
		t.Fail()
	}
}
