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
		creator := players.APlayer("player-1", "Hermione")
		store := NewInMemoryGameStore(nil, nil)
		err := store.AddPendingGame(gameID, creator)
		utils.AssertNoError(t, err)

		ps, ok := store.FindPendingPlayers(gameID)
		utils.AssertTrue(t, ok)
		utils.AssertEqual(t, ps[0], creator)
	})

	t.Run("Can retrieve existing game", func(t *testing.T) {
		id := "test-game-id"

		game, _ := New(id, players.SomePlayers(), nil)

		store := NewInMemoryGameStore(
			map[string]GameEngine{
				id: game,
			},
			nil,
		)

		game, ok := store.FindActiveGame(id)
		utils.AssertTrue(t, ok)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindActiveGame("fake-id")

		utils.AssertEqual(t, ok, false)
	})

	t.Run("Can retrieve pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			map[string]players.Players{pendingID: players.SomePlayers()})

		game, ok := store.FindPendingPlayers(pendingID)

		utils.AssertTrue(t, ok)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindPendingPlayers("fake-id")
		utils.AssertEqual(t, ok, false)
	})

	t.Run("Can add a player to a pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			map[string]players.Players{pendingID: players.SomePlayers()},
		)

		playerToAdd := players.APlayer("player-id", "Horatio")
		err := store.AddToPendingPlayers(pendingID, playerToAdd)
		utils.AssertNoError(t, err)

		pendingPlayers, ok := store.FindPendingPlayers(pendingID)
		utils.AssertTrue(t, ok)

		got, ok := pendingPlayers.Find(playerToAdd.ID)
		utils.AssertTrue(t, ok)
		utils.AssertEqual(t, got, playerToAdd)
	})
}
