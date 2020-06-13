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
	players   Players
	stage     Stage
	deck      deck.Deck
}

// New constructs a new GameEngine
func New(players []*Player) (*GameEngine, error) {
	if len(players) < 2 {
		return nil, fmt.Errorf("Could not construct Game: minimum of 2 players required (supplied %d)", len(players))
	}
	if len(players) > 4 {
		return nil, fmt.Errorf("Could not construct Game: maximum of 4 players allowed (supplied %d)", len(players))
	}

	engine := GameEngine{
		players: Players(players),
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
	messages []OutboundMessage,
) (
	[]InboundMessage,
	error,
) {
	ch := make(chan InboundMessage)
	// TODO: make less rubbish
	for _, m := range messages {
		for _, p := range ge.players {
			if p.ID == m.PlayerID {
				go p.sendMessageAwaitReply(ch, m)
				break // debug
			}
		}
	}

	responses := []InboundMessage{}
	for i := 0; i < len(messages); i++ {
		resp := <-ch
		responses = append(responses, resp)
	}

	return responses, nil
}
