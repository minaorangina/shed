package gameengine

import "github.com/minaorangina/shed/player"

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
