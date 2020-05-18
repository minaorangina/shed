package gameengine

import (
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func initialisedGameEngine() GameEngine {
	playerNames := []string{"Ada", "Katherine"}
	engine := New()
	engine.Init(playerNames)
	return engine
}

func TestGameEngineInit(t *testing.T) {
	ge := initialisedGameEngine()

	// produces a game
	if ge.game == nil {
		t.Fatal("GameEngine.game is nil")
	}

	// produces ExternalPlayers
	if ge.externalPlayers == nil {
		t.Fatal("GameEngine.externalPlayers is nil")
	}
	if len(ge.externalPlayers) != 2 {
		t.Errorf(utils.FailureMessage(2, len(ge.externalPlayers)))
	}

	// correct playState
	type playStateTest struct {
		testName   string
		gameEngine GameEngine
		expected   playState
	}
	playStateTests := []playStateTest{
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

	for _, test := range playStateTests {
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
	engine := initialisedGameEngine()
	game := *engine.game
	game.start() // deal cards

	messages := make(map[string]messageToPlayer)
	expected := map[string]messageFromPlayer{}
	for _, p := range game.players {
		o := buildOpponents(p.id, game.players)
		m := game.buildMessageToPlayer(p, o, "Rearrange your initial cards")
		messages[p.id] = m

		expected[p.id] = messageFromPlayer{
			PlayerID:  p.id,
			Hand: p.hand,
			Seen: p.seen,
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
