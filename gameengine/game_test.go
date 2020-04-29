package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestGame(t *testing.T) {
	gameEngine, _ := New([]string{"harry-1", "sally-1"})
	player1, err := NewPlayer("harry-1", "Harry")
	if err != nil {
		t.Errorf(err.Error())
	}
	player2, err := NewPlayer("sally-1", "Sally")
	if err != nil {
		t.Errorf(err.Error())
	}
	somePlayers := []Player{player1, player2}
	_ = somePlayers

	game := NewGame(&gameEngine, []playerInfo{{"Harry-1", "Harry"}, {"Sally-1", "Sally"}})
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
		c := p.cards()
		numHand := len(c.hand)
		numSeen := len(c.seen)
		numUnseen := len(c.unseen)
		if numHand != 3 {
			formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
			t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
		}
	}

	// buildOpponents
	player0, _ := NewPlayer("hermy-0", "Hermione")
	someMorePlayers := []Player{player0, player2}
	expectedOpponents := []opponent{
		{ID: player0.id, Name: player0.name, SeenCards: player0.cards().seen},
		{ID: player2.id, Name: player2.name, SeenCards: player2.cards().seen},
	}
	opponents := buildOpponents("harry-1", someMorePlayers)
	if !reflect.DeepEqual(opponents, expectedOpponents) {
		t.Errorf("\nExpected: %+v\nActual: %+v\n", expectedOpponents, opponents)
	}

	// buildMessageToPlayer
	playerToContact := (*game.players)[1]
	message := game.buildMessageToPlayer(playerToContact, opponents, "Let the games begin!")
	expectedMessage := messageToPlayer{
		Message:   "Let the games begin!",
		PlayState: gameEngine.playState,
		GameStage: game.stage,
		PlayerID:  playerToContact.id,
		HandCards: playerToContact.cards().hand,
		SeenCards: playerToContact.cards().seen,
		Opponents: expectedOpponents,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		t.Errorf(utils.FailureMessage(utils.TypeToString(expectedMessage), utils.TypeToString(message)))
	}
}
