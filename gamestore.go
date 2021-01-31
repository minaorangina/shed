package shed

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownGameID           = errors.New("unknown game ID")
	ErrUnknownPlayerID         = errors.New("unknown player ID")
	ErrFnUnknownInactiveGameID = func(gameID string) error {
		return fmt.Errorf("pending game with id \"%s\" does not exist", gameID)
	}
	ErrGameAlreadyStarted = errors.New("game has already started")
)

type PlayerInfo struct {
	PlayerID, Name string
}

type GameStore interface {
	FindGame(gameID string) GameEngine
	FindActiveGame(gameID string) GameEngine
	FindInactiveGame(gameID string) GameEngine
	FindPendingPlayer(gameID, playerID string) *PlayerInfo
	AddInactiveGame(engine GameEngine) error
	AddPendingPlayer(gameID, playerID, name string) error
	AddPlayerToGame(gameID string, player Player) error
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	Games          map[string]GameEngine
	PendingPlayers map[string][]PlayerInfo
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore() *InMemoryGameStore {
	return &InMemoryGameStore{
		Games:          map[string]GameEngine{},
		PendingPlayers: map[string][]PlayerInfo{},
	}
}

func (s *InMemoryGameStore) FindGame(ID string) GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}

	return game
}

// does this need a mutex?
func (s *InMemoryGameStore) FindActiveGame(ID string) GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}
	if game.PlayState() == Idle {
		return nil
	}
	return game
}

func (s *InMemoryGameStore) FindInactiveGame(ID string) GameEngine {
	game, ok := s.Games[ID]
	if !ok {
		return nil
	}
	// to be replaced by something real
	if game.PlayState() != Idle {
		return nil
	}
	return game
}

func (s *InMemoryGameStore) FindPendingPlayer(gameID, playerID string) *PlayerInfo {
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
func (s *InMemoryGameStore) AddInactiveGame(game GameEngine) error {
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

	if game.PlayState() != Idle {
		return ErrGameAlreadyStarted
	}

	// mutex required
	s.PendingPlayers[gameID] = append(s.PendingPlayers[gameID], PlayerInfo{playerID, name})

	return nil
}

func (s *InMemoryGameStore) AddPlayerToGame(gameID string, player Player) error {
	game := s.FindInactiveGame(gameID)
	if game == nil {
		return ErrFnUnknownInactiveGameID(gameID)
	}

	return game.AddPlayer(player)
}
