package gameengine

import (
	"fmt"
)

// playState represents the state of the current game
// idle -> no game play (pre game and post game)
// inProgress -> game in progress
// paused -> game is paused
type playState int

func (gps playState) String() string {
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
	idle playState = iota
	inProgress
	paused
)

// GameEngine represents the engine of the game
type GameEngine struct {
	playState   playState
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
	if ge.playState != idle {
		return nil
	}
	// new game
	shedGame := NewGame(ge, ge.playerNames)
	ge.game = shedGame

	ge.playState = inProgress
	ge.game.start()

	return nil
}

// PlayState returns the current gameplay status
func (ge *GameEngine) PlayState() string {
	return ge.playState.String()
}

func (ge *GameEngine) start() {
	ge.playState = inProgress
}
