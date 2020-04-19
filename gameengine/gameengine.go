package gameengine

import (
	"fmt"
)

// gameState represents the state of the current game
// idle -> no game play (pre game and post game)
// inProgress -> game in progress
// paused -> game is paused
type gameState int

func (gps gameState) String() string {
	if gps == 0 {
		return "idle"
	} else if gps == 1 {
		return "inProgress"
	} else if gps == 2 {
		return "paused"
	}
	return ""
}

const (
	idle gameState = iota
	inProgress
	paused
)

// GameEngine represents the engine of the game
type GameEngine struct {
	gameState   gameState
	playerNames []string
	game        *Game
}

// New constructs a new GameEngine
func New(playerNames []string) (GameEngine, error) {
	if len(playerNames) < 2 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: minimum of 2 players required (supplied %d)", len(playerNames))
	}
	if len(playerNames) > 4 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: maximum of 4 players allowed (supplied %d)", len(playerNames))
	}

	return GameEngine{playerNames: playerNames}, nil
}

// Init initialises a new game
func (ge *GameEngine) Init() error {
	if ge.gameState != idle {
		return nil
	}
	// new game
	shedGame := NewGame(ge.playerNames)
	ge.game = shedGame

	ge.gameState = inProgress
	ge.game.start()

	return nil
}

// GameState returns the current gameplay status
func (ge *GameEngine) GameState() string {
	return ge.gameState.String()
}

func (ge *GameEngine) start() {
	ge.gameState = inProgress
}
