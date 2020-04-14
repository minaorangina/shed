package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/minaorangina/shed/player"
)

func TestGame(t *testing.T) {
	player1, _ := player.New("Harry")
	player2, _ := player.New("Sally")
	somePlayers := []player.Player{player1, player2}

	game := NewGame(somePlayers)
	expectedGame := Game{"Shed", somePlayers}
	if !reflect.DeepEqual(expectedGame, game) {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %v\n", expectedGame, game))
	}
}
