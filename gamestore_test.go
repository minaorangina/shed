package shed

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
)

func TestInMemoryGameStore(t *testing.T) {
	t.Run("Constructor prevents nil struct members", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		if store.ActiveGames() == nil {
			t.Error("gamestore.game was nil")
		}
		if store.PendingGames() == nil {
			t.Error("gamestore.pending was nil")
		}
	})

	t.Run("Can create a new pending game", func(t *testing.T) {
		gameID := "some-game-id"
		playerID := "player-1"
		creator := players.APlayer(playerID, "Hermione")

		store := NewInMemoryGameStore(nil, nil)
		err := store.AddPendingGame(gameID, creator)
		utils.AssertNoError(t, err)

		game, ok := store.FindPendingGame(gameID)
		utils.AssertTrue(t, ok)

		ps := game.Players()
		_, ok = ps.Find(playerID)
		utils.AssertTrue(t, ok)
	})

	t.Run("Can retrieve existing active game", func(t *testing.T) {
		id := "test-game-id"

		store := NewInMemoryGameStore(
			newActiveGame(id, players.SomePlayers()),
			nil,
		)

		game, ok := store.FindActiveGame(id)
		utils.AssertTrue(t, ok)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent active game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindActiveGame("fake-id")

		utils.AssertEqual(t, ok, false)
	})

	t.Run("Can retrieve existing pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			newPendingGame(pendingID, players.SomePlayers()),
		)

		game, ok := store.FindPendingGame(pendingID)

		utils.AssertTrue(t, ok)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindPendingGame("fake-id")
		utils.AssertEqual(t, ok, false)
	})

	t.Run("Can add a player to a pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			newPendingGame(pendingID, players.SomePlayers()),
		)

		playerToAdd := players.APlayer("player-id", "Horatio")
		err := store.AddToPendingPlayers(pendingID, playerToAdd)
		utils.AssertNoError(t, err)

		pendingGame, ok := store.FindPendingGame(pendingID)
		utils.AssertTrue(t, ok)

		pendingPlayers := pendingGame.Players()
		got, ok := pendingPlayers.Find(playerToAdd.ID)
		utils.AssertTrue(t, ok)
		utils.AssertEqual(t, got, playerToAdd)
	})

	t.Run("Disallows adding a player to an active game", func(t *testing.T) {
		gameID := "test-game-id"
		creator := players.NewPlayers(players.APlayer("some-player-id", "Horatio"))

		store := NewInMemoryGameStore(
			newActiveGame(gameID, creator),
			nil,
		)

		playerToAdd := players.APlayer("player-1", "Neville")
		err := store.AddToPendingPlayers(gameID, playerToAdd)

		utils.AssertErrored(t, err)
	})

	t.Run("Can activate a pending game", func(t *testing.T) {
		gameID := "some-game-ID"

		store := NewInMemoryGameStore(
			nil,
			newPendingGame(gameID, players.SomePlayers()),
		)

		err := store.ActivateGame(gameID)
		utils.AssertNoError(t, err)

		_, ok := store.FindActiveGame(gameID)
		utils.AssertTrue(t, ok)

		_, ok = store.FindPendingGame(gameID)
		utils.AssertEqual(t, ok, false)
	})

	t.Run("Activating an already-active game is a no-op", func(t *testing.T) {
		gameID := "this-is-a-game-id"

		store := NewInMemoryGameStore(
			newActiveGame(gameID, nil),
			nil,
		)

		err := store.ActivateGame(gameID)
		utils.AssertNoError(t, err)
	})
}

func newActiveGame(gameID string, ps players.Players) map[string]GameEngine {
	game, _ := New(gameID, ps, nil)
	return map[string]GameEngine{gameID: game}
}

func newPendingGame(gameID string, ps players.Players) map[string]GameEngine {
	game, _ := New(gameID, ps, nil)
	return map[string]GameEngine{gameID: game}
}
