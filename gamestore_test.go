package shed

import (
	"testing"

	"github.com/minaorangina/shed/players"
)

func TestInMemoryGameStore(t *testing.T) {
	id := "test-game-id"
	pendingID := "pending-game"

	game, _ := New(id, SomePlayers(), nil)

	store := InMemoryGameStore{
		games: map[string]GameEngine{
			id: game,
		},
		pending: map[string]players.Players{
			pendingID: SomePlayers(),
		},
	}

	t.Run("Can retrieve existing game", func(t *testing.T) {
		game, ok := store.GetGame(id)
		AssertTrue(t, ok)
		if game == nil {
			t.Error("Game was nil")
		}
	})

	t.Run("Handles a non-existent game", func(t *testing.T) {
		_, ok := store.GetGame("fake-id")
		AssertEqual(t, ok, false)
	})

	t.Run("Can retrieve pending game", func(t *testing.T) {
		game, ok := store.GetPendingGame(pendingID)
		AssertTrue(t, ok)
		if game == nil {
			t.Error("Game was nil")
		}
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		_, ok := store.GetPendingGame("fake-id")
		AssertEqual(t, ok, false)
	})
}
