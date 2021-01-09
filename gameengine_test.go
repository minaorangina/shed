package shed

import (
	"fmt"
	"testing"
	"time"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

const gameEngineTestTimeout = time.Duration(200 * time.Millisecond)

func TestGameEngineConstructor(t *testing.T) {
	creatorID := "hermione-1"
	t.Run("keeps track of who created it", func(t *testing.T) {
		engine, err := NewGameEngine(GameEngineOpts{GameID: "some-id", CreatorID: creatorID, Game: &SpyGame{}})
		utils.AssertNoError(t, err)
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

		engine, err := NewGameEngine(GameEngineOpts{GameID: "game-id", CreatorID: player1ID, Game: &SpyGame{}})
		utils.AssertNoError(t, err)

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

		engine, err := NewGameEngine(GameEngineOpts{
			GameID:       "game-id",
			CreatorID:    player1ID,
			Players:      players,
			UnregisterCh: unregisterCh,
			Game:         &SpyGame{},
		})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, engine)

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
	t.Run("has an ID", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		engine, err := NewGameEngine(GameEngineOpts{
			GameID:    gameID,
			CreatorID: playerID,
			Players:   SomePlayers(),
			Game:      &SpyGame{}})
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, engine.ID(), gameID)
	})

	t.Run("has the user ID of the creator", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		engine, err := NewGameEngine(GameEngineOpts{
			GameID:    gameID,
			CreatorID: playerID,
			Players:   SomePlayers(),
			Game:      &SpyGame{}})
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, engine.CreatorID(), playerID)
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
			{
				"just right",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy"}),
				nil,
			},
		}

		for _, et := range testsShouldError {
			ge, err := NewGameEngine(GameEngineOpts{Players: et.input, Game: NewShed(ShedOpts{})})
			utils.AssertNoError(t, err)
			utils.AssertNotNil(t, ge)
			err = ge.Start()
			utils.AssertDeepEqual(t, err, et.want)
		}
	})

	t.Run("starting more than once is a no-op", func(t *testing.T) {
		engine, err := NewGameEngine(GameEngineOpts{Players: SomePlayers(), Game: &SpyGame{}})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, engine)

		engine.playState = InProgress
		err = engine.Start()
		utils.AssertNoError(t, err)
	})
}

func TestGameEngineReceiveMessage(t *testing.T) {
	t.Run("performs correct action for start game command", func(t *testing.T) {
		spy := &SpyGame{}
		engine, err := NewGameEngine(GameEngineOpts{Players: SomePlayers(), Game: spy})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, engine)

		msg := InboundMessage{
			PlayerID: "an-id",
			Command:  protocol.Start,
		}
		engine.Receive(msg)

		utils.Within(t, gameEngineTestTimeout, func() {
			utils.AssertTrue(t, spy.startCalled)
		})

		utils.AssertEqual(t, engine.playState, InProgress)
	})
}
