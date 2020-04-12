package gameengine

import (
	"fmt"

	"github.com/minaorangina/shed/player"
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

// stage represents the main stages in the game
type stage int

const (
	handOrganisation stage = iota
	clearDeck
	clearHand
)

// GameEngine represents the engine of the game
type GameEngine struct {
	gameState gameState
	// should player be in its own package?
	players []player.Player
	stage   stage
}

// New constructs a new GameEngine
func New(playerNames []string) (GameEngine, error) {
	if len(playerNames) < 2 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: minimum of 2 players required (supplied %d)", len(playerNames))
	}
	if len(playerNames) > 4 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: maximum of 4 players allowed (supplied %d)", len(playerNames))
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

// Init initialises a new game
func (ge *GameEngine) Init() error {
	if ge.gameState != idle {
		return fmt.Errorf("Cannot call `GameEngine.Init()` when game is not idle (currently %s)", ge.gameState.String())
	}
	ge.start()

	return nil
}

// GameState returns the current gameplay status
func (ge *GameEngine) GameState() string {
	return ge.gameState.String()
}

func (ge *GameEngine) start() {
	ge.gameState = inProgress
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
