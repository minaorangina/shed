package gameengine

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGame(t *testing.T) {
	gameEngine, _ := New([]string{"Harry", "Sally"})
	player1, err := NewPlayer(1, "Harry")
	if err != nil {
		t.Errorf(err.Error())
	}
	player2, err := NewPlayer(2, "Sally")
	if err != nil {
		t.Errorf(err.Error())
	}
	somePlayers := []Player{player1, player2}
	_ = somePlayers

	game := NewGame(&gameEngine, []string{"Harry", "Sally"})
	if len(game.deck) != 52 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 52, len(game.deck)))
	}
	if len(*game.players) != 2 {
		t.Errorf(fmt.Sprintf("\nExpected: %+v\nActual: %+v\n", 2, len(*game.players)))
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

	// buildOpponents
	player0, _ := NewPlayer(0, "Hermione")
	someMorePlayers := []Player{player0, player1, player2}
	expectedOpponents := []opponent{{ID: 0, Name: "Hermione"}, {ID: 2, Name: "Sally"}}
	opponents := buildOpponents(1, someMorePlayers)
	if !reflect.DeepEqual(opponents, expectedOpponents) {
		t.Errorf("\nExpected: %+v\nActual: %+v\n", expectedOpponents, opponents)
	}
}
