package engine

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/minaorangina/shed/game"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

const gameEngineTestTimeout = time.Duration(200 * time.Millisecond)

func TestGameEngineConstructor(t *testing.T) {
	creatorID := "hermione-1"
	t.Run("keeps track of who created it", func(t *testing.T) {
		ge, err := NewGameEngine(GameEngineOpts{GameID: "some-id", CreatorID: creatorID, Game: &SpyGame{}})
		utils.AssertNoError(t, err)
		utils.AssertEqual(t, ge.CreatorID(), creatorID)
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

		ge, err := NewGameEngine(GameEngineOpts{GameID: "game-id", CreatorID: player1ID, Game: &SpyGame{}})
		utils.AssertNoError(t, err)

		player1.ge = ge
		ge.players = NewPlayers(player1)

		joiningPlayer := APlayer("joiner-1", "Ms Joiner")
		ge.AddPlayer(joiningPlayer)

		utils.Within(t, gameEngineTestTimeout, func() {
			msg, ok := <-sendCh
			utils.AssertTrue(t, ok)

			var data protocol.OutboundMessage
			err := json.Unmarshal(msg, &data)
			utils.AssertNoError(t, err)
			fmt.Println(data)
			utils.AssertEqual(t, data.Joiner.Name, joiningPlayer.Name())
			utils.AssertEqual(t, data.Joiner.PlayerID, joiningPlayer.ID())
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

		ge, err := NewGameEngine(GameEngineOpts{
			GameID:       "game-id",
			CreatorID:    player1ID,
			Players:      players,
			UnregisterCh: unregisterCh,
			Game:         &SpyGame{},
		})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, ge)

		go func() {
			unregisterCh <- player1
		}()

		utils.Within(t, gameEngineTestTimeout, func() {
			_, ok := <-unregisterCh
			utils.AssertTrue(t, ok)
			utils.AssertEqual(t, len(ge.players), 1)
		})
	})
}

func TestGameEngineInit(t *testing.T) {
	t.Run("has an ID", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		ge, err := NewGameEngine(GameEngineOpts{
			GameID:    gameID,
			CreatorID: playerID,
			Players:   SomePlayers(),
			Game:      &SpyGame{}})
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, ge.ID(), gameID)
	})

	t.Run("has the user ID of the creator", func(t *testing.T) {
		gameID := "thisistheid"
		playerID := "i created it"
		ge, err := NewGameEngine(GameEngineOpts{
			GameID:    gameID,
			CreatorID: playerID,
			Players:   SomePlayers(),
			Game:      &SpyGame{}})
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, ge.CreatorID(), playerID)
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
				game.ErrTooFewPlayers,
			},
			{
				"too many players",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
				game.ErrTooManyPlayers,
			},
			{
				"just right",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy"}),
				nil,
			},
		}

		for _, et := range testsShouldError {
			ge, err := NewGameEngine(GameEngineOpts{Players: et.input, Game: game.ExistingShed(game.ShedOpts{})})
			utils.AssertNoError(t, err)
			utils.AssertNotNil(t, ge)
			err = ge.Start()
			utils.AssertDeepEqual(t, err, et.want)
		}
	})

	t.Run("starting more than once is a no-op", func(t *testing.T) {
		ge, err := NewGameEngine(GameEngineOpts{Players: SomePlayers(), Game: &SpyGame{}})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, ge)

		ge.playState = InProgress
		err = ge.Start()
		utils.AssertNoError(t, err)
	})
}

func TestGameEngineReceiveMessage(t *testing.T) {
	t.Run("performs correct action for start game command", func(t *testing.T) {
		spy := NewSpyGame()
		ge, err := NewGameEngine(GameEngineOpts{Players: SomePlayers(), Game: spy})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, ge)

		msg := protocol.InboundMessage{
			PlayerID: "an-id",
			Command:  protocol.Start,
		}
		ge.Receive(msg)

		utils.Within(t, gameEngineTestTimeout, func() {
			utils.AssertTrue(t, spy.startCalled)
		})

		utils.AssertEqual(t, ge.playState, InProgress)
	})
}
