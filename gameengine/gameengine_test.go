package gameengine

import (
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestGameEngineInit(t *testing.T) {
	type test struct {
		testName   string
		gameEngine GameEngine
		expected   playState
	}
	initTests := []test{
		{
			testName:   "`Init` puts game engine into inProgress state",
			gameEngine: GameEngine{},
			expected:   inProgress,
		},
		{
			testName: "`Init` does nothing if game in progress",
			gameEngine: GameEngine{
				playState: inProgress,
			},
			expected: inProgress,
		},
		{
			testName: "`Init` does nothing if game paused",
			gameEngine: GameEngine{
				playState: paused,
			},
			expected: paused,
		},
	}

	for _, test := range initTests {
		playerNames := []string{"Ada", "Katherine"}
		err := test.gameEngine.Init(playerNames)
		if err != nil {
			t.Fatalf("Failed to intialise game: %s", err.Error())
		}
		if test.expected != test.gameEngine.playState {
			t.Errorf(utils.TableFailureMessage(test.testName, test.expected.String(), test.gameEngine.playState.String()))
		}
	}
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// Game Engine receives from messages to send to players
	// and returns response
	playerNames := []string{"Ada", "Katherine"}
	engine := New()
	err := engine.Init(playerNames)
	if err != nil {
		t.Fatalf("Failed to intialise game engine: %s", err.Error())
	}

	game := *engine.game
	game.start() // deal cards

	messages := make(map[string]messageToPlayer)
	expected := map[string]reorganisedHand{}
	for _, p := range *game.players {
		o := buildOpponents(p.id, *game.players)
		m := game.buildMessageToPlayer(p, o, "Rearrange your hand")
		messages[p.id] = m

		expected[p.id] = reorganisedHand{
			PlayerID:  p.id,
			HandCards: p.hand,
			SeenCards: p.seen,
		}
	}

	actual, err := engine.messagePlayersAwaitReply(messages)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(utils.FailureMessage(expected, actual))
	}
}
