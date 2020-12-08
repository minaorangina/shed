package shed

import (
	"errors"
	"fmt"
	"testing"
	"time"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

var gameEngineTestTimeout = time.Duration(200 * time.Millisecond)

type spySetup struct {
	called bool
}

func (s *spySetup) setup(ge GameEngine) error {
	s.called = true
	return nil
}

func TestGameEngineConstructor(t *testing.T) {
	creatorID := "hermione-1"
	t.Run("keeps track of who created it", func(t *testing.T) {
		engine := NewGameEngine(GameEngineOpts{GameID: "some-id", CreatorID: creatorID})
		utils.AssertEqual(t, engine.CreatorID(), creatorID)
	})
}

func TestGameEngineAddPlayer(t *testing.T) {
	t.Run("can add more players", func(t *testing.T) {
		playerID := "some-player-id"
		name := "Héloise"
		ge := gameEngineWithPlayers()

		err := ge.AddPlayer(APlayer(playerID, name))
		utils.AssertNoError(t, err)

		ps := ge.Players()
		_, ok := ps.Find(playerID)
		utils.AssertTrue(t, ok)
	})

	t.Run("broadcasts to other players", func(t *testing.T) {
		sendCh := make(chan []byte)
		player1ID := "i-am-a-spy"

		player1 := &WSPlayer{
			id:     player1ID,
			name:   "Spy",
			conn:   nil,
			sendCh: sendCh,
		}

		engine := NewGameEngine(GameEngineOpts{GameID: "game-id", CreatorID: player1ID})

		player1.engine = engine
		engine.players = NewPlayers(player1)

		joiningPlayer := APlayer("joiner-1", "Ms Joiner")
		engine.AddPlayer(joiningPlayer)

		utils.Within(t, gameEngineTestTimeout, func() {
			msg, ok := <-sendCh
			utils.AssertTrue(t, ok)
			utils.AssertEqual(t, fmt.Sprintf("%s has joined the game!", joiningPlayer.Name()), string(msg))
		})
	})
}

func TestRemovePlayer(t *testing.T) {
	t.Run("unregisters players", func(t *testing.T) {
		t.Skip()
		unregisterCh := make(chan Player)
		player1ID, player2ID := "itsame", "itsaalsome"

		player1 := &WSPlayer{
			id:   player1ID,
			name: "Mario",
			conn: nil,
		}
		player2 := &WSPlayer{
			id:   player2ID,
			name: "Luigi",
			conn: nil,
		}

		players := NewPlayers(player1, player2)

		engine := NewGameEngine(GameEngineOpts{GameID: "game-id", CreatorID: player1ID, Players: players, UnregisterCh: unregisterCh})

		go func() {
			unregisterCh <- player1
		}()

		utils.Within(t, gameEngineTestTimeout, func() {
			_, ok := <-unregisterCh
			utils.AssertTrue(t, ok)
			utils.AssertEqual(t, len(engine.players), 1)
		})
	})
}
func TestGameEngineInit(t *testing.T) {
	t.Run("constructs with correct number of cards", func(t *testing.T) {
		ge := gameEngineWithPlayers()
		if len(ge.Deck()) != 52 {
			t.Errorf("\nExpected: %+v\nActual: %+v\n", 52, len(ge.Deck()))
		}
	})
	t.Run("has an ID", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		engine := NewGameEngine(GameEngineOpts{GameID: gameID, CreatorID: playerID, Players: SomePlayers()})

		utils.AssertEqual(t, engine.ID(), gameID)
	})

	t.Run("has the user ID of the creator", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		engine := NewGameEngine(GameEngineOpts{GameID: gameID, CreatorID: playerID, Players: SomePlayers()})

		utils.AssertEqual(t, engine.CreatorID(), playerID)
	})
}
func TestGameEngineSetupFn(t *testing.T) {
	t.Run("does not error if no setup fn defined", func(t *testing.T) {
		engine := NewGameEngine(GameEngineOpts{Players: SomePlayers()})

		err := engine.Setup()
		utils.AssertNoError(t, err)
	})

	t.Run("propagates setup fn error", func(t *testing.T) {
		erroringSetupFn := func(ge GameEngine) error {
			return errors.New("Whoops")
		}
		engine := NewGameEngine(GameEngineOpts{Players: SomePlayers(), SetupFn: erroringSetupFn})

		err := engine.Setup()
		if err == nil {
			t.Fatalf("Expected an error, but there was none")
		}
	})
}

func TestGameEngineStart(t *testing.T) {
	t.Run("only starts with legal number of players", func(t *testing.T) {
		type gameTest struct {
			testName string
			input    Players
			want     error
		}
		testsShouldError := []gameTest{
			{
				"too few players",
				namesToPlayers([]string{"Grace"}),
				ErrTooFewPlayers,
			},
			{
				"too many players",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
				ErrTooManyPlayers,
			},
		}

		for _, et := range testsShouldError {
			ge := NewGameEngine(GameEngineOpts{Players: et.input})
			err := ge.Start()
			utils.AssertEqual(t, err.Error(), et.want.Error())
		}
	})

	t.Run("starting more than once is a no-op", func(t *testing.T) {
		engine := NewGameEngine(GameEngineOpts{Players: SomePlayers()})

		engine.playState = InProgress
		err := engine.Start()
		utils.AssertNoError(t, err)
	})

	t.Run("calls the setup fn", func(t *testing.T) {
		spy := spySetup{}
		engine := NewGameEngine(GameEngineOpts{Players: SomePlayers(), SetupFn: spy.setup})

		err := engine.Start()
		utils.AssertNoError(t, err)

		if spy.called != true {
			t.Errorf("Expected spy setup fn to be called")
		}
	})
}

func TestGameEngineReceiveMessage(t *testing.T) {
	t.Run("performs correct action for start game command", func(t *testing.T) {
		var called bool
		setupFn := func(ge GameEngine) error {
			called = true
			return nil
		}

		engine := NewGameEngine(GameEngineOpts{Players: SomePlayers(), SetupFn: setupFn})

		msg := InboundMessage{
			PlayerID: "an-id",
			Command:  protocol.Start,
		}
		engine.Receive(msg)

		utils.Within(t, gameEngineTestTimeout, func() {
			utils.AssertTrue(t, called)
		})

		utils.AssertEqual(t, engine.playState, InProgress)
	})
}
