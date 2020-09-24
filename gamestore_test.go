package shed

import (
	"testing"

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
		creator := APlayer("player-1", "Hermione")
		store := NewInMemoryGameStore(nil, nil)
		err := store.AddPendingGame(gameID, creator)
		AssertNoError(t, err)

		ps, ok := store.FindPendingPlayers(gameID)
		AssertTrue(t, ok)
		AssertEqual(t, ps[0], creator)
	})

	t.Run("Can retrieve existing game", func(t *testing.T) {
		id := "test-game-id"

		game, _ := New(id, SomePlayers(), nil)

		store := NewInMemoryGameStore(
			map[string]GameEngine{
				id: game,
			},
			nil,
		)

		game, ok := store.FindActiveGame(id)
		AssertTrue(t, ok)
		AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindActiveGame("fake-id")

		AssertEqual(t, ok, false)
	})

	t.Run("Can retrieve pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			map[string]players.Players{pendingID: SomePlayers()})

		game, ok := store.FindPendingPlayers(pendingID)

		AssertTrue(t, ok)
		AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		store := NewInMemoryGameStore(nil, nil)
		_, ok := store.FindPendingPlayers("fake-id")
		AssertEqual(t, ok, false)
	})

	t.Run("Can add a player to a pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewInMemoryGameStore(
			nil,
			map[string]players.Players{pendingID: SomePlayers()},
		)

		playerToAdd := APlayer("player-id", "Horatio")
		err := store.AddToPendingPlayers(pendingID, playerToAdd)
		AssertNoError(t, err)

		pendingPlayers, ok := store.FindPendingPlayers(pendingID)
		AssertTrue(t, ok)

		got, ok := pendingPlayers.Find(playerToAdd.ID)
		AssertTrue(t, ok)
		AssertEqual(t, got, playerToAdd)
	})
}
