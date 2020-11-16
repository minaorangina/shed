package shed

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestInMemoryGameStore(t *testing.T) {
	t.Run("Constructor prevents nil struct members", func(t *testing.T) {
		store := NewInMemoryGameStore()
		if store.ActiveGames == nil {
			t.Error("Active games was nil")
		}
		if store.InactiveGames == nil {
			t.Error("Pending games was nil")
		}
		if store.PendingPlayers == nil {
			t.Error("Pending players was nil")
		}
	})

	t.Run("Can create a new pending game", func(t *testing.T) {

	})

	t.Run("Can add pending players", func(t *testing.T) {
		gameID := "some-game-id"
		playerID, playerName := "player-1", "Hermione"
		engine, err := New(gameID, playerID, nil, nil)
		utils.AssertNoError(t, err)

		store := NewTestGameStore(
			nil,
			map[string]GameEngine{
				gameID: engine,
			},
			nil,
		)

		err = store.AddPendingPlayer(gameID, playerID, playerName)
		utils.AssertNoError(t, err)

		pendingInfo := store.FindPendingPlayer(gameID, playerID)
		utils.AssertNotNil(t, pendingInfo)
	})

	t.Run("Can retrieve existing active game", func(t *testing.T) {
		gameID := "test-game-id"

		store := NewTestGameStore(
			newActiveGame(gameID, "", SomePlayers()),
			nil, nil,
		)

		game := store.FindActiveGame(gameID)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent active game", func(t *testing.T) {
		store := NewInMemoryGameStore()
		game := store.FindActiveGame("fake-id")

		utils.AssertEqual(t, game, nil)
	})

	t.Run("Can retrieve existing pending game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := &InMemoryGameStore{
			ActiveGames:    map[string]GameEngine{},
			InactiveGames:  NewInactiveGame(pendingID, "creator-id", SomePlayers()),
			PendingPlayers: map[string][]PlayerInfo{},
		}

		game := store.FindInactiveGame(pendingID)
		utils.AssertNotNil(t, game)
	})

	t.Run("Handles a non-existent pending game", func(t *testing.T) {
		store := NewInMemoryGameStore()
		game := store.FindInactiveGame("fake-id")
		utils.AssertEqual(t, game, nil)
	})

	t.Run("Can add a player to an inactive game", func(t *testing.T) {
		pendingID := "a-pending-game"
		store := NewTestGameStore(
			nil,
			NewInactiveGame(pendingID, "creator-id", SomePlayers()),
			nil,
		)

		playerID, playerName := "horatio-1", "Horatio"
		playerToAdd := APlayer(playerID, playerName)

		err := store.AddPlayerToGame(pendingID, playerToAdd)
		utils.AssertNoError(t, err)

		game := store.FindInactiveGame(pendingID)
		utils.AssertNotNil(t, game)

		ps := game.Players()
		p, ok := ps.Find(playerID)

		utils.AssertTrue(t, ok)
		utils.AssertEqual(t, p, playerToAdd)
	})

	t.Run("Disallows adding a player to an active game", func(t *testing.T) {
		gameID := "test-game-id"
		creator := NewPlayers(APlayer("some-player-id", "Horatio"))

		store := NewTestGameStore(
			newActiveGame(gameID, "creator-id", creator),
			nil,
			nil,
		)

		playerID, playerName := "player-1", "Neville"
		err := store.AddPendingPlayer(gameID, playerID, playerName)

		utils.AssertErrored(t, err)
	})

	t.Run("Can activate a pending game", func(t *testing.T) {
		gameID := "some-game-ID"

		store := NewTestGameStore(
			nil,
			NewInactiveGame(gameID, "creator-id", SomePlayers()),
			nil,
		)

		err := store.ActivateGame(gameID)
		utils.AssertNoError(t, err)

		game := store.FindActiveGame(gameID)
		utils.AssertNotNil(t, game)

		game = store.FindInactiveGame(gameID)
		utils.AssertEqual(t, game, nil)
	})

	t.Run("Activating an already-active game is a no-op", func(t *testing.T) {
		gameID := "this-is-a-game-id"

		store := NewTestGameStore(
			newActiveGame(gameID, "creator-id", nil),
			nil,
			nil,
		)

		err := store.ActivateGame(gameID)
		utils.AssertNoError(t, err)
	})
}

func newActiveGame(gameID, playerID string, ps Players) map[string]GameEngine {
	game, _ := New(gameID, playerID, ps, nil)
	return map[string]GameEngine{gameID: game}
}

func NewInactiveGame(gameID, playerID string, ps Players) map[string]GameEngine {
	game, _ := New(gameID, playerID, ps, nil)
	return map[string]GameEngine{gameID: game}
}
