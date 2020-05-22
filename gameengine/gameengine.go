package gameengine

import (
	"fmt"

	"github.com/minaorangina/shed/deck"
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

type playerInfo struct {
	id   string
	name string
}

// GameEngine represents the engine of the game
type GameEngine struct {
	playState playState
	players   AllPlayers
	stage     Stage
	deck      deck.Deck
}

// New constructs a new GameEngine
func New(players AllPlayers) (*GameEngine, error) {
	if len(players) < 2 {
		return nil, fmt.Errorf("Could not construct Game: minimum of 2 players required (supplied %d)", len(players))
	}
	if len(players) > 4 {
		return nil, fmt.Errorf("Could not construct Game: maximum of 4 players allowed (supplied %d)", len(players))
	}

	engine := GameEngine{
		players: players,
		deck:    deck.New(),
	}

	return &engine, nil
}

// Start starts a game
func (ge *GameEngine) Start() error {
	if ge.playState != idle {
		return nil
	}

	ge.playState = inProgress

	err := ge.handleInitialCards() // mock?
	if err != nil {
		return err
	}

	// play actual game
	return nil
}

func (ge *GameEngine) messagePlayersAwaitReply(
	messages OutboundMessages,
) (
	InboundMessages,
	error,
) {
	ch := make(chan messageFromPlayer)
	for _, p := range ge.players {
		go p.sendMessageAwaitReply(ch, messages[p.ID])
		break // debug
	}

	responses := InboundMessages{}
	for i := 0; i < len(messages); i++ {
		resp := <-ch
		responses.Add(resp.PlayerID, resp)
	}

	return responses, nil
}
