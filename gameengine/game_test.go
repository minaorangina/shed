package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/minaorangina/shed/deck"
)

func TestGame(t *testing.T) {
	gameEngine, _ := New([]string{"Harry", "Sally"})
	player1, _ := NewPlayer("Harry")
	player2, _ := NewPlayer("Sally")
	somePlayers := []Player{player1, player2}

	game := NewGame(&gameEngine, []string{"Harry", "Sally"})
	expectedGame := Game{Name: "Shed", players: &somePlayers, deck: deck.New(), engine: &gameEngine}
	if !reflect.DeepEqual(expectedGame, *game) {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", expectedGame, game))
	}

	expectedStage := "handOrganisation"
	if game.Stage() != expectedStage {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", expectedStage, game.Stage()))
	}

	game.start()

	if game.players == nil {
		t.Fatal("game.player is nil, which was not expected")
	}

	for _, p := range *game.players {
		c := p.cards
		numHand := len(*c.hand)
		numSeen := len(*p.cards.seen)
		numUnseen := len(*p.cards.unseen)
		if numHand != 3 {
			formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
			t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
		}
	}
}
