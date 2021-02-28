package store

import (
	"errors"
	"fmt"

	"github.com/minaorangina/shed/engine"
	"github.com/minaorangina/shed/protocol"
)

var (
	ErrUnknownGameID           = errors.New("unknown game ID")
	ErrUnknownPlayerID         = errors.New("unknown player ID")
	ErrFnUnknownInactiveGameID = func(gameID string) error {
		return fmt.Errorf("pending game with id \"%s\" does not exist", gameID)
	}
	ErrGameAlreadyStarted = errors.New("game has already started")
)

type GameStore interface {
	FindGame(gameID string) engine.GameEngine
	FindActiveGame(gameID string) engine.GameEngine
	FindInactiveGame(gameID string) engine.GameEngine
	FindPendingPlayer(gameID, playerID string) *protocol.PlayerInfo
	AddInactiveGame(engine engine.GameEngine) error
	AddPendingPlayer(gameID, playerID, name string) error
	AddPlayerToGame(gameID string, player engine.Player) error
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	Games          map[string]engine.GameEngine
	PendingPlayers map[string][]protocol.PlayerInfo
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore() *InMemoryGameStore {
	return &InMemoryGameStore{
		Games:          map[string]engine.GameEngine{},
		PendingPlayers: map[string][]protocol.PlayerInfo{},
	}
}

func (s *InMemoryGameStore) FindGame(ID string) engine.GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}

	return game
}

// does this need a mutex?
func (s *InMemoryGameStore) FindActiveGame(ID string) engine.GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}
	if game.PlayState() == engine.Idle {
		return nil
	}
	return game
}

func (s *InMemoryGameStore) FindInactiveGame(ID string) engine.GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}
	// to be replaced by something real
	if game.PlayState() != engine.Idle {
		return nil
	}
	return game
}

func (s *InMemoryGameStore) FindPendingPlayer(gameID, playerID string) *protocol.PlayerInfo {
	pendingPlayers, ok := s.PendingPlayers[gameID]
	if !ok {
		return nil
	}

	for i, info := range pendingPlayers {
		if info.PlayerID == playerID {
			return &pendingPlayers[i]
		}
	}

	return nil
}

// mutex definitely required
func (s *InMemoryGameStore) AddInactiveGame(game engine.GameEngine) error {
	if game, exists := s.Games[game.ID()]; exists {
		return fmt.Errorf("Game with id %s already exists", game.ID())
	}

	s.Games[game.ID()] = game
	return nil
}

// AddPendingPlayer adds the information from which to construct a Player in the future.
// If the target Game does not exist, it will fail.
func (s *InMemoryGameStore) AddPendingPlayer(gameID, playerID, name string) error {
	game := s.FindInactiveGame(gameID)
	if game == nil {
		return ErrFnUnknownInactiveGameID(gameID)
	}

	if game.PlayState() != engine.Idle {
		return ErrGameAlreadyStarted
	}

	// mutex required
	s.PendingPlayers[gameID] = append(s.PendingPlayers[gameID], protocol.PlayerInfo{playerID, name})

	return nil
}

func (s *InMemoryGameStore) AddPlayerToGame(gameID string, player engine.Player) error {
	game := s.FindInactiveGame(gameID)
	if game == nil {
		return ErrFnUnknownInactiveGameID(gameID)
	}

	return game.AddPlayer(player)
}
