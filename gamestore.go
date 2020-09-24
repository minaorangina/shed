package shed

import (
	"fmt"

	"github.com/minaorangina/shed/players"
)

type GameStore interface {
	ActiveGames() map[string]GameEngine
	PendingGames() map[string]players.Players
	FindActiveGame(ID string) (GameEngine, bool)
	FindPendingPlayers(ID string) (players.Players, bool)
	AddPendingGame(ID string, creator *players.Player) error
	AddToPendingPlayers(ID string, player *players.Player) error
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	active  map[string]GameEngine
	pending map[string]players.Players
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore(
	active map[string]GameEngine,
	pending map[string]players.Players,
) GameStore {
	if active == nil {
		active = map[string]GameEngine{}
	}
	if pending == nil {
		pending = map[string]players.Players{}
	}
	return &InMemoryGameStore{active, pending}
}

func (s *InMemoryGameStore) ActiveGames() map[string]GameEngine {
	return s.active
}
func (s *InMemoryGameStore) PendingGames() map[string]players.Players {
	return s.pending
}

// does this need a mutex?
func (s *InMemoryGameStore) FindActiveGame(ID string) (GameEngine, bool) {
	game, ok := s.active[ID]
	return game, ok
}

func (s *InMemoryGameStore) FindPendingPlayers(ID string) (players.Players, bool) {
	pendingPlayers, ok := s.pending[ID]
	return pendingPlayers, ok
}

// mutex definitely required
func (s *InMemoryGameStore) AddPendingGame(gameID string, creator *players.Player) error {
	if _, exists := s.pending[gameID]; exists {
		return fmt.Errorf("Game with id %s already exists", gameID)
	}
	s.pending[gameID] = players.NewPlayers(creator)
	return nil
}

func (s *InMemoryGameStore) AddToPendingPlayers(ID string, player *players.Player) error {
	pendingPlayers, ok := s.pending[ID]
	if !ok {
		return fmt.Errorf("Pending game with id %s does not exist", ID)
	}

	s.pending[ID] = players.AddPlayer(&pendingPlayers, player)

	return nil
}
