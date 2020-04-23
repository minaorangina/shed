package gameengine

import (
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p, err := NewPlayer(1, name)

	if err != nil {
		t.Errorf(err.Error())
	}
	pc := playerCards{}
	if !reflect.DeepEqual(Player{1, name, &pc}, p) {
		t.Fail()
	}
}
