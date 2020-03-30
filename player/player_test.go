package player

import (
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p, err := New(name)

	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(Player{name}, p) {
		t.Fail()
	}
}
