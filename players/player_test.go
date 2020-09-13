package players

import (
	"os"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p := NewPlayer("player-1", name, os.Stdin, os.Stdout)
	if p.Conn == nil {
		t.Fail()
	}
}

// func TestNewPlayer(t *testing.T) {
// 	SomePlayers()
// }
