package gameengine

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestInit(t *testing.T) {
	expectedPlayerNames := []string{"Ada", "Katherine"}

	type test struct {
		testName   string
		gameEngine GameEngine
		expected   playState
	}
	initNoopTests := []test{
		{
			testName:   "`Init` puts game engine into inProgress state",
			gameEngine: GameEngine{playerNames: expectedPlayerNames},
			expected:   inProgress,
		},
		{
			testName: "`Init` does nothing if game in progress",
			gameEngine: GameEngine{
				playerNames: expectedPlayerNames,
				playState:   inProgress,
			},
			expected: inProgress,
		},
		{
			testName: "`Init` does nothing if game paused",
			gameEngine: GameEngine{
				playerNames: expectedPlayerNames,
				playState:   paused,
			},
			expected: paused,
		},
	}

	for _, test := range initNoopTests {
		err := test.gameEngine.Init()
		if err != nil {
			t.Errorf(utils.TableFailureMessage(test.testName, "[no error]", err.Error()))
		}
		if test.expected != test.gameEngine.playState {
			t.Errorf(utils.TableFailureMessage(test.testName, test.expected.String(), test.gameEngine.playState.String()))
		}
	}
}