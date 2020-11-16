package shed

import (
	"errors"
	"fmt"

	"github.com/minaorangina/shed/players"
)

var (
	ErrUnknownGameID          = errors.New("unknown game ID")
	ErrUnknownPlayerID        = errors.New("unknown player ID")
	ErrFnUnknownPendingGameID = func(gameID string) error {
		return fmt.Errorf("pending game with id \"%s\" does not exist", gameID)
	}
)

type PlayerInfo struct {
	userID, name string
}

type GameStore interface {
	FindActiveGame(gameID string) GameEngine
	FindInactiveGame(gameID string) GameEngine
	FindPendingPlayer(gameID, userID string) *PlayerInfo
	AddInactiveGame(engine GameEngine) error
	AddPendingPlayer(gameID, userID, name string) error
	AddPlayerToGame(gameID string, player players.Player) error
	ActivateGame(gameID string) error
}

// InMemoryGameStore maps game id to game engine
type InMemoryGameStore struct {
	ActiveGames    map[string]GameEngine
	PendingPlayers map[string][]PlayerInfo
	InactiveGames  map[string]GameEngine
}

// NewInMemoryGameStore constructs an InMemoryGameStore
func NewInMemoryGameStore() *InMemoryGameStore {
	return &InMemoryGameStore{
		ActiveGames:    map[string]GameEngine{},
		InactiveGames:  map[string]GameEngine{},
		PendingPlayers: map[string][]PlayerInfo{},
	}
}

// does this need a mutex?
func (s *InMemoryGameStore) FindActiveGame(ID string) GameEngine {
	return s.ActiveGames[ID]
}

func (s *InMemoryGameStore) FindInactiveGame(ID string) GameEngine {
	return s.InactiveGames[ID]
}

func (s *InMemoryGameStore) FindPendingPlayer(gameID, userID string) *PlayerInfo {
	pendingPlayers, ok := s.PendingPlayers[gameID]
	if !ok {
		return nil
	}

	for i, info := range pendingPlayers {
		if info.userID == userID {
			return &pendingPlayers[i]
		}
	}

	return nil
}

// mutex definitely required
func (s *InMemoryGameStore) AddInactiveGame(game GameEngine) error {
	if _, exists := s.InactiveGames[game.ID()]; exists {
		return fmt.Errorf("Game with id %s already exists", game.ID())
	}

	s.InactiveGames[game.ID()] = game

	return nil
}

// AddPendingPlayer adds the information from which to construct a Player in the future.
// If the target Game does not exist, it will fail.
func (s *InMemoryGameStore) AddPendingPlayer(gameID, userID, name string) error {
	game := s.FindInactiveGame(gameID)
	if game == nil {
		return ErrFnUnknownPendingGameID(gameID)
	}

	// mutex required
	s.PendingPlayers[gameID] = append(s.PendingPlayers[gameID], PlayerInfo{userID, name})

	return nil
}

func (s *InMemoryGameStore) AddPlayerToGame(gameID string, player players.Player) error {
	game := s.FindInactiveGame(gameID)
	if game == nil {
		return ErrFnUnknownPendingGameID(gameID)
	}

	return game.AddPlayer(player)
}

func (s *InMemoryGameStore) ActivateGame(gameID string) error {
	activeGame := s.FindActiveGame(gameID)
	if activeGame != nil {
		return nil
	}

	inactiveGame := s.FindInactiveGame(gameID)
	if inactiveGame == nil {
		return ErrFnUnknownPendingGameID(gameID)
	}

	// needs mutex
	s.ActiveGames[gameID] = inactiveGame
	delete(s.InactiveGames, gameID)

	return nil
}
