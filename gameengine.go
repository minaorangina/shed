package shed

import (
	"errors"

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
type GameEngine interface {
	Setup() error
	Start() error
	MessagePlayers([]players.OutboundMessage) ([]players.InboundMessage, error)
	Deck() deck.Deck
	Players() players.Players
	ID() string
	AddPlayer(*players.Player) error
}

type gameEngine struct {
	id        string
	playState playState
	players   players.Players
	stage     Stage
	deck      deck.Deck
	setupFn   func(GameEngine) error
}

var (
	ErrTooFewPlayers  = errors.New("minimum of 2 players required")
	ErrTooManyPlayers = errors.New("maximum of 4 players allowed")
)

// New constructs a new GameEngine
func New(id string, players players.Players, setupFn func(GameEngine) error) (GameEngine, error) {
	engine := gameEngine{
		id:      id,
		players: players,
		deck:    deck.New(),
		setupFn: setupFn,
	}

	return &engine, nil
}

func (ge *gameEngine) ID() string {
	return ge.id
}

// Setup does any pre-game setup required
func (ge *gameEngine) Setup() error {
	if err := ge.CheckNumPlayers(); err != nil {
		return err
	}

	var err error
	if ge.setupFn != nil {
		err = ge.setupFn(ge)
	}
	return err
}

// AddPlayer adds a player to a game
func (ge *gameEngine) AddPlayer(p *players.Player) error {
	ps := ge.Players()
	ge.players = players.AddPlayer(&ps, p)
	return nil
}

// Start starts a game
// Might be renamed `next`
func (ge *gameEngine) Start() error {
	if err := ge.CheckNumPlayers(); err != nil {
		return err
	}

	if ge.playState != idle {
		return nil
	}

	ge.playState = inProgress
	// next tick?
	return nil
}

func (ge *gameEngine) CheckNumPlayers() error {
	if len(ge.players) < 2 {
		return ErrTooFewPlayers
	}
	if len(ge.players) > 4 {
		return ErrTooManyPlayers
	}

	return nil
}

func (ge *gameEngine) MessagePlayers(messages []players.OutboundMessage) ([]players.InboundMessage, error) {
	return messagePlayersAwaitReply(ge.Players(), messages)
}

func (ge *gameEngine) Deck() deck.Deck {
	return ge.deck
}

func (ge *gameEngine) Players() players.Players {
	// mutex?
	return ge.players
}

func messagePlayersAwaitReply(
	ps players.Players,
	messages []players.OutboundMessage,
) (
	[]players.InboundMessage,
	error,
) {
	ch := make(chan players.InboundMessage)
	for _, m := range messages {
		if p, ok := ps.Find(m.PlayerID); ok {
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
