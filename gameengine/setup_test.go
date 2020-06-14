package gameengine

import (
	"fmt"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestSetupFn(t *testing.T) {
	t.Run("all cards dealt", func(t *testing.T) {
		t.Skip("do not run TestGameEngineStart/all_cards_dealt")
		ge, _ := gameEngineWithPlayers()
		err := ge.Start() // mock required
		if err != nil {
			t.Fatal("Unexpected error ", err.Error())
		}
		for _, p := range ge.players {
			c := p.Cards()
			numHand := len(c.Hand)
			numSeen := len(c.Seen)
			numUnseen := len(c.Unseen)
			if numHand != 3 {
				formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
				t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
			}
		}
	})

	t.Run("correct playState", func(t *testing.T) {
		t.Skip("skip testing playstates")
		type playStateTest struct {
			testName   string
			gameEngine GameEngine
			expected   playState
		}
		playStateTests := []playStateTest{
			{
				testName:   "`Start` puts game engine into inProgress state",
				gameEngine: GameEngine{},
				expected:   inProgress,
			},
			{
				testName: "`Start` does nothing if game in progress",
				gameEngine: GameEngine{
					playState: inProgress,
				},
				expected: inProgress,
			},
			{
				testName: "`Start` does nothing if game paused",
				gameEngine: GameEngine{
					playState: paused,
				},
				expected: paused,
			},
		}

		for _, test := range playStateTests {
			err := test.gameEngine.Start()
			if err != nil {
				t.Fatalf("Failed to intialise game: %s", err.Error())
			}
			if test.expected != test.gameEngine.playState {
				utils.TableFailureMessage(t, test.testName, test.expected.String(), test.gameEngine.playState.String())
			}
		}
	})
}
