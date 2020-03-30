package gameengine

import (
	"fmt"

	"github.com/minaorangina/shed/player"
)

// gamePlayStatus represents the status of the current game
// idle -> no game play (pre game and post game)
// inProgress -> game in progress
// paused -> game is paused
type gamePlayStatus int

const (
	idle gamePlayStatus = iota
	inProgress
	paused
)

// stage represents the main stages in the game
type stage int

const (
	handOrganisation stage = iota
	clearCardQueue
	clearHand
)

// GameEngine represents the engine of the game
type GameEngine struct {
	gameplayStatus gamePlayStatus
	// should player be in its own package?
	players []player.Player
	stage   stage
}

// New constructs a new GameEngine
func New(playerNames []string) (GameEngine, error) {
	if len(playerNames) < 2 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: minimum 2 players required (supplied %d)", len(playerNames))
	}
	players, err := namesToPlayers(playerNames)
	if err != nil {
		return GameEngine{}, err
	}

	engine := GameEngine{
		players: players,
	}

	return engine, nil
}

func namesToPlayers(names []string) ([]player.Player, error) {
	players := make([]player.Player, 0, len(names))
	for _, name := range names {
		p, playerErr := player.New(name)
		if playerErr != nil {
			return []player.Player{}, playerErr
		}
		players = append(players, p)
	}

	return players, nil
}
