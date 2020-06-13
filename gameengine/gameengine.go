package gameengine

import (
	"fmt"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/players"
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
	players   players.Players
	stage     Stage
	deck      deck.Deck
}

// New constructs a new GameEngine
func New(players []*players.Player) (*GameEngine, error) {
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
	messages []players.OutboundMessage,
) (
	[]players.InboundMessage,
	error,
) {
	ch := make(chan players.InboundMessage)
	for _, m := range messages {
		if p, ok := ge.players.Individual(m.PlayerID); ok {
			go p.SendMessageAwaitReply(ch, m)
			break // debug
		}
	}

	responses := []players.InboundMessage{}
	for i := 0; i < len(messages); i++ {
		resp := <-ch
		responses = append(responses, resp)
	}

	return responses, nil
}
