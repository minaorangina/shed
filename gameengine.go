package shed

import (
	"errors"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
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
type GameEngine interface {
	Setup() error
	Start() error
	MessagePlayers([]OutboundMessage) error
	Deck() deck.Deck
	Players() Players
	ID() string
	CreatorID() string
	AddPlayer(Player) error
}

// gameEngine represents the engine of the game

type gameEngine struct {
	id          string
	creatorID   string
	playState   playState
	players     Players
	registerCh  chan Player
	broadcastCh chan []byte
	stage       Stage
	deck        deck.Deck
	setupFn     func(GameEngine) error
}

var (
	ErrTooFewPlayers  = errors.New("minimum of 2 players required")
	ErrTooManyPlayers = errors.New("maximum of 4 players allowed")
)

// New constructs a new GameEngine
func NewGameEngine(gameID string, creatorID string, players Players, setupFn func(GameEngine) error) (*gameEngine, error) {
	engine := &gameEngine{
		id:          gameID,
		creatorID:   creatorID,
		players:     players,
		registerCh:  make(chan Player),
		broadcastCh: make(chan []byte),
		deck:        deck.New(),
		setupFn:     setupFn,
	}

	// Listen for websocket connections
	go engine.Listen()

	return engine, nil
}

func (ge *gameEngine) ID() string {
	return ge.id
}

func (ge *gameEngine) CreatorID() string {
	return ge.creatorID
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
func (ge *gameEngine) AddPlayer(p Player) error {
	ge.registerCh <- p
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

func (ge *gameEngine) MessagePlayers(messages []OutboundMessage) error {
	return messagePlayersAwaitReply(ge.Players(), messages)
}

func (ge *gameEngine) Deck() deck.Deck {
	return ge.deck
}

func (ge *gameEngine) Players() Players {
	// mutex?
	return ge.players
}

// outside of the interface
func messagePlayersAwaitReply(
	ps Players,
	messages []OutboundMessage,
) error {
	for _, m := range messages {
		if p, ok := ps.Find(m.PlayerID); ok {
			go p.Send(m)
			break // debug
		}
	}

	// responses := []InboundMessage{}
	// for i := 0; i < len(messages); i++ {
	// 	resp := <-ch
	// 	responses = append(responses, resp)
	// }

	return nil
}

// outside of the interface
func (ge *gameEngine) Listen() {
	for {
		select {
		case player := <-ge.registerCh:
			ps := ge.Players()
			ge.players = AddPlayer(ps, player)
			ge.broadcast([]byte(player.Name()))

		case msg := <-ge.broadcastCh:
			for _, p := range ge.players {
				outbound := buildNewJoinerMessage(p, msg)
				p.Send(outbound)
			}
		}
	}
}

// outside of the interface
func (ge *gameEngine) broadcast(msg []byte) {
	ge.broadcastCh <- msg
}

func buildNewJoinerMessage(player Player, msg []byte) OutboundMessage {
	return OutboundMessage{
		PlayerID:  player.ID(),
		Name:      player.Name(),
		Message:   string(msg),
		Hand:      nil,
		Seen:      nil,
		Opponents: nil,
		Command:   protocol.NewJoiner,
	}
}
