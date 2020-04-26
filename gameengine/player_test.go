package gameengine

import (
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p, err := NewPlayer("player-1", name)

	if err != nil {
		t.Errorf(err.Error())
	}
	expectedPlayer := Player{
		name: name,
		id:   "player-1",
	}
	if !reflect.DeepEqual(expectedPlayer, p) {
		t.Fail()
	}
}
