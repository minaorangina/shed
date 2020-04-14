package gameengine

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/player"
)

func TestInit(t *testing.T) {
	expectedPlayers := []player.Player{{Name: "Ada"}, {Name: "Katherine"}}

	type test struct {
		testName   string
		gameEngine GameEngine
		expected   gameState
	}
	initNoopTests := []test{
		{
			testName:   "`Init` puts game engine into inProgress state",
			gameEngine: GameEngine{players: expectedPlayers},
			expected:   inProgress,
		},
		{
			testName: "`Init` does nothing if game in progress",
			gameEngine: GameEngine{
				players:   expectedPlayers,
				gameState: inProgress,
			},
			expected: inProgress,
		},
		{
			testName: "`Init` does nothing if game paused",
			gameEngine: GameEngine{
				players:   expectedPlayers,
				gameState: paused,
			},
			expected: paused,
		},
	}

	for _, test := range initNoopTests {
		err := test.gameEngine.Init()
		if err != nil {
			t.Errorf(utils.TableFailureMessage(test.testName, "[no error]", err.Error()))
		}
		if test.expected != test.gameEngine.gameState {
			t.Errorf(utils.TableFailureMessage(test.testName, test.expected.String(), test.gameEngine.gameState.String()))
		}
	}
}
