package shed

import "github.com/minaorangina/shed/players"

type GameStore interface {
	GetGame(ID string) (GameEngine, bool)
	GetPendingGame(ID string) (players.Players, bool)
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	games   map[string]GameEngine
	pending map[string]players.Players
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore(
	games map[string]GameEngine,
	pending map[string]players.Players,
) *InMemoryGameStore {
	return &InMemoryGameStore{games, pending}
}

// does this need a mutex?
func (s *InMemoryGameStore) GetGame(ID string) (GameEngine, bool) {
	game, ok := s.games[ID]
	return game, ok
}

func (s *InMemoryGameStore) GetPendingGame(ID string) (players.Players, bool) {
	pendingPlayers, ok := s.pending[ID]
	return pendingPlayers, ok
}
