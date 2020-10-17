package shed

import (
	"fmt"

	"github.com/minaorangina/shed/players"
)

type GameStore interface {
	ActiveGames() map[string]GameEngine
	PendingGames() map[string]GameEngine
	FindActiveGame(ID string) (GameEngine, bool)
	FindPendingGame(ID string) (GameEngine, bool)
	AddPendingGame(ID string, creator *players.Player) error
	AddToPendingPlayers(ID string, player *players.Player) error
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	active  map[string]GameEngine
	pending map[string]GameEngine
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore(
	active map[string]GameEngine,
	pending map[string]GameEngine,
) GameStore {
	if active == nil {
		active = map[string]GameEngine{}
	}
	if pending == nil {
		pending = map[string]GameEngine{}
	}
	return &InMemoryGameStore{active, pending}
}

func (s *InMemoryGameStore) ActiveGames() map[string]GameEngine {
	return s.active
}
func (s *InMemoryGameStore) PendingGames() map[string]GameEngine {
	return s.pending
}

// does this need a mutex?
func (s *InMemoryGameStore) FindActiveGame(ID string) (GameEngine, bool) {
	game, ok := s.active[ID]
	return game, ok
}

func (s *InMemoryGameStore) FindPendingGame(ID string) (GameEngine, bool) {
	pendingGame, ok := s.pending[ID]
	return pendingGame, ok
}

// mutex definitely required
func (s *InMemoryGameStore) AddPendingGame(gameID string, creator *players.Player) error {
	if _, exists := s.pending[gameID]; exists {
		return fmt.Errorf("Game with id %s already exists", gameID)
	}
	game, err := New(gameID, players.NewPlayers(creator), nil)
	if err != nil {
		return err
	}

	s.pending[gameID] = game

	return nil
}

func (s *InMemoryGameStore) AddToPendingPlayers(ID string, player *players.Player) error {
	pendingGame, ok := s.FindPendingGame(ID)
	if !ok {
		return fmt.Errorf("Pending game with id %s does not exist", ID)
	}

	// does this need to be reassigned to the map?
	err := pendingGame.AddPlayer(player)

	return err
}
