package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/minaorangina/shed/deck"
)

func TestGame(t *testing.T) {
	player1, _ := NewPlayer("Harry")
	player2, _ := NewPlayer("Sally")
	somePlayers := []Player{player1, player2}

	game := NewGame([]string{"Harry", "Sally"})
	expectedGame := Game{"Shed", &somePlayers, deck.New()}
	if !reflect.DeepEqual(expectedGame, *game) {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %v\n", expectedGame, game))
	}
}
