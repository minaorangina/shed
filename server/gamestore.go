package server

import "github.com/minaorangina/shed"

type GameStore interface {
	GetGame(ID string) (shed.GameEngine, bool)
}

// inMemoryGameStore maps game id to game engine
type inMemoryGameStore struct {
	games map[string]shed.GameEngine
}

// does this need a mutex?
func (s *inMemoryGameStore) GetGame(ID string) (shed.GameEngine, bool) {
	game, ok := s.games[ID]
	return game, ok
}
