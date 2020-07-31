package server

import ge "github.com/minaorangina/shed/gameengine"

type GameStore interface {
	GetGame(ID string) (ge.GameEngine, bool)
}

type inMemoryGameStore struct {
	games map[string]ge.GameEngine
}

func (s *inMemoryGameStore) GetGame(ID string) (ge.GameEngine, bool) {
	game, ok := s.games[ID]
	return game, ok
}
