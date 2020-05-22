package gameengine

import (
	"fmt"
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

// HANGS
func TestGameEngineStart(t *testing.T) {
	t.Run("produces players", func(t *testing.T) {
		ge, _ := gameEngineWithPlayers()
		// produces Players
		if ge.players == nil {
			t.Fatal("GameEngine.players is nil")
		}
		if len(ge.players) != 2 {
			t.Errorf(utils.FailureMessage(2, len(ge.players)))
		}
	})

	t.Run("starts correctly", func(t *testing.T) {
		t.Skip("do not run TestGameEngineStart/starts_correctly")
		ge, _ := gameEngineWithPlayers()
		err := ge.Start() // mock required
		if err != nil {
			t.Fatalf("Could not start game")
		}
	})

	t.Run("all cards dealt", func(t *testing.T) {
		t.Skip("do not run TestGameEngineStart/all_cards_dealt")
		ge, _ := gameEngineWithPlayers()
		err := ge.Start() // mock required
		if err != nil {
			t.Fatal("Unexpected error ", err.Error())
		}
		for _, p := range ge.players {
			c := p.cards()
			numHand := len(c.hand)
			numSeen := len(c.seen)
			numUnseen := len(c.unseen)
			if numHand != 3 {
				formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
				t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
			}
		}
	})

	t.Run("correct playState", func(t *testing.T) {
		// correct playState
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
				t.Errorf(utils.TableFailureMessage(test.testName, test.expected.String(), test.gameEngine.playState.String()))
			}
		}
	})
}

// DO NOT RUN
func TestGameEngineMsgFromGame(t *testing.T) {
	// Game Engine receives from messages to send to players
	// and returns response
	t.Skip("do not run TestGameEngineMsgFromGame")
	ge, _ := gameEngineWithPlayers()
	ge.Start() // mock required

	messages := OutboundMessages{}
	expected := InboundMessages{}
	initialCards := initialCards{}
	for _, p := range ge.players {
		o := buildOpponents(p.ID, ge.players)
		m := ge.buildReorgMessage(p, o, initialCards, "Rearrange your initial cards")
		messages.Add(p.ID, m)

		expected[p.ID] = messageFromPlayer{
			PlayerID: p.ID,
			Hand:     p.hand,
			Seen:     p.seen,
		}
	}

	actual, err := ge.messagePlayersAwaitReply(messages)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(utils.FailureMessage(expected, actual))
	}
}
