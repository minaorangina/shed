package store

import (
	"testing"

	"github.com/minaorangina/shed/engine"
	"github.com/minaorangina/shed/game"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

func TestInMemoryGameStore(t *testing.T) {
	t.Run("Constructor prevents nil struct members", func(t *testing.T) {
		str := NewInMemoryGameStore()
		if str.Games == nil {
			t.Error("Games was nil")
		}
		if str.PendingPlayers == nil {
			t.Error("Pending players was nil")
		}
	})

	t.Run("prevents duplicate game IDs", func(t *testing.T) {
		str := NewInMemoryGameStore()
		gameID := "thisISAnID"
		ge, _ := engine.NewGameEngine(engine.GameEngineOpts{GameID: gameID, Game: &engine.SpyGame{}})

		err := str.AddInactiveGame(ge)
		utils.AssertNoError(t, err)

		err = str.AddInactiveGame(ge)
		utils.AssertErrored(t, err)
	})

	t.Run("Can add pending players", func(t *testing.T) {
		gameID := "some-game-id"
		playerID, playerName := "player-1", "Hermione"
		ge, err := engine.NewGameEngine(engine.GameEngineOpts{GameID: gameID, CreatorID: playerID, Game: &engine.SpyGame{}})
		utils.AssertNoError(t, err)
		utils.AssertNotNil(t, ge)

		str := NewTestGameStore(
			map[string]engine.GameEngine{
				gameID: ge,
			},
			nil,
		)

		err = str.AddPendingPlayer(gameID, playerID, playerName)
		utils.AssertNoError(t, err)

		pendingInfo := str.FindPendingPlayer(gameID, playerID)
		utils.AssertNotNil(t, pendingInfo)
	})

	t.Run("Handles a non-existent game", func(t *testing.T) {
		str := NewInMemoryGameStore()
		game := str.FindGame("fake-id")

		utils.AssertEqual(t, game, nil)
	})

	t.Run("Can add a player to an inactive game", func(t *testing.T) {
		pendingID := "a-pending-game"
		str := NewTestGameStore(
			NewInactiveGame(pendingID, "creator-id", engine.SomePlayers()),
			nil,
		)

		playerID, playerName := "horatio-1", "Horatio"
		playerToAdd := engine.APlayer(playerID, playerName)

		err := str.AddPlayerToGame(pendingID, playerToAdd)
		utils.AssertNoError(t, err)

		game := str.FindInactiveGame(pendingID)
		utils.AssertNotNil(t, game)

		ps := game.Players()
		p, ok := ps.Find(playerID)

		utils.AssertTrue(t, ok)
		utils.AssertEqual(t, p, playerToAdd)
	})

	t.Run("Disallows adding a player to an active game", func(t *testing.T) {
		gameID := "test-game-id"
		creator := engine.NewPlayers(engine.APlayer("some-player-id", "Horatio"))

		str := NewTestGameStore(
			newActiveGame(gameID, "creator-id", creator),
			nil,
		)

		playerID, playerName := "player-1", "Neville"
		err := str.AddPendingPlayer(gameID, playerID, playerName)

		utils.AssertErrored(t, err)
	})

	// May remove distinction between active and inactive games one day...

	t.Run("Can retrieve existing active game", func(t *testing.T) {
		gameID := "test-game-id"

		str := NewTestGameStore(
			newActiveGame(gameID, "", engine.SomePlayers()),
			nil,
		)

		game := str.FindActiveGame(gameID)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent active game", func(t *testing.T) {
		str := NewInMemoryGameStore()
		game := str.FindActiveGame("fake-id")

		utils.AssertEqual(t, game, nil)
	})

	t.Run("Can retrieve existing pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		str := &InMemoryGameStore{
			Games:          NewInactiveGame(pendingID, "creator-id", engine.SomePlayers()),
			PendingPlayers: map[string][]protocol.Player{},
		}

		game := str.FindInactiveGame(pendingID)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		str := NewInMemoryGameStore()
		game := str.FindInactiveGame("fake-id")
		utils.AssertEqual(t, game, nil)
	})
}

func newActiveGame(gameID, playerID string, ps engine.Players) map[string]engine.GameEngine {
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    gameID,
		CreatorID: playerID,
		Players:   ps,
		PlayState: engine.InProgress,
		Game:      game.ExistingShed(game.ShedOpts{}),
	})
	return map[string]engine.GameEngine{gameID: game}
}

func NewInactiveGame(gameID, playerID string, ps engine.Players) map[string]engine.GameEngine {
	game, _ := engine.NewGameEngine(engine.GameEngineOpts{
		GameID:    gameID,
		CreatorID: playerID,
		Players:   ps,
		Game:      game.ExistingShed(game.ShedOpts{}),
	})
	return map[string]engine.GameEngine{gameID: game}
}

// NewTestGameStore is a convenience function for creating InMemoryGameStore in tests
func NewTestGameStore(
	games map[string]engine.GameEngine,
	pendingPlayers map[string][]protocol.Player,
) *InMemoryGameStore {
	if games == nil {
		games = map[string]engine.GameEngine{}
	}

	if pendingPlayers == nil {
		pendingPlayers = map[string][]protocol.Player{}
	}

	return &InMemoryGameStore{
		Games:          games,
		PendingPlayers: pendingPlayers,
	}
}
