package gameengine

import (
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p, err := NewPlayer(name)

	if err != nil {
		t.Errorf(err.Error())
	}
	pc := playerCards{}
	if !reflect.DeepEqual(Player{name, &pc}, p) {
		t.Fail()
	}
}
