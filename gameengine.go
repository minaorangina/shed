package shed

import (
	"errors"
	"fmt"

	"github.com/minaorangina/shed/protocol"
)

var (
	ErrNilGame                = errors.New("game is nil")
	ErrTooFewPlayers          = errors.New("minimum of 2 players required")
	ErrTooManyPlayers         = errors.New("maximum of 4 players allowed")
	ErrNoPlayers              = errors.New("game has no players")
	ErrGameUnexpectedResponse = errors.New("game received unexpected response")
	ErrGameAwaitingResponse   = errors.New("game is awaiting a response")
)

// PlayState represents the state of the current game
// idle -> no game play (pre game and post game)
// InProgress -> game in progress
// paused -> game is paused
type PlayState int

func (gps PlayState) String() string {
	if gps == 0 {
		return "idle"
	} else if gps == 1 {
		return "InProgress"
	} else if gps == 2 {
		return "paused"
	}
	return ""
}

const (
	Idle PlayState = iota
	InProgress
	Paused
)

// GameEngine represents the engine of the game
type GameEngine interface {
	Start() error
	MessagePlayers([]OutboundMessage) error
	Players() Players
	ID() string
	CreatorID() string
	AddPlayer(Player) error
	RemovePlayer(Player)
	Receive(InboundMessage)
	PlayState() PlayState
}

// gameEngine represents the engine of the game

type gameEngine struct {
	id           string
	creatorID    string
	playState    PlayState
	players      Players
	registerCh   chan Player
	unregisterCh chan Player
	inboundCh    chan InboundMessage
	game         Game
}

// GameEngineOpts represents options for constructing a new GameEngine
type GameEngineOpts struct {
	GameID                   string
	CreatorID                string
	Players                  Players
	RegisterCh, UnregisterCh chan Player
	InboundCh                chan InboundMessage
	PlayState                PlayState
	Game                     Game
}

// NewGameEngine constructs a new GameEngine
func NewGameEngine(opts GameEngineOpts) (*gameEngine, error) {
	if opts.Game == nil {
		return nil, ErrNilGame
	}
	if opts.RegisterCh == nil {
		opts.RegisterCh = make(chan Player)
	}
	if opts.InboundCh == nil {
		opts.InboundCh = make(chan InboundMessage)
	}
	engine := &gameEngine{
		id:           opts.GameID,
		creatorID:    opts.CreatorID,
		players:      opts.Players,
		registerCh:   opts.RegisterCh,
		unregisterCh: opts.UnregisterCh,
		inboundCh:    opts.InboundCh,
		playState:    opts.PlayState,
		game:         opts.Game,
	}

	// Listen for websocket connections
	go engine.Listen()

	return engine, nil
}

// AddPlayer adds a player to a game
func (ge *gameEngine) AddPlayer(p Player) error {
	if ge.playState != Idle {
		return errors.New("cannot add player - game has started")
	}
	ge.registerCh <- p
	return nil
}

func (ge *gameEngine) RemovePlayer(p Player) {
	ge.unregisterCh <- p
}

// Start starts a game
// Might be renamed `next`
func (ge *gameEngine) Start() error {
	if ge.playState != Idle {
		return nil
	}

	if ge.game == nil {
		return ErrNilGame
	}

	// until there's a proper way to get ids
	ids := []string{}
	for _, p := range ge.players {
		ids = append(ids, p.ID())
	}

	err := ge.game.Start(ids)
	if err != nil {
		return err
	}

	// mutex
	ge.playState = InProgress

	return nil
}

func (ge *gameEngine) MessagePlayers(messages []OutboundMessage) error {
	missingPlayers := []string{}
	for _, m := range messages {
		p, ok := ge.players.Find(m.PlayerID)
		if !ok {
			missingPlayers = append(missingPlayers, m.PlayerID)
		}
		p.Send(m)
	}
	if len(missingPlayers) > 0 {
		return fmt.Errorf("could not send to some players")
	}

	return nil
}

// Receive forwards InboundMessages from Players for sorting
func (ge *gameEngine) Receive(msg InboundMessage) {
	ge.inboundCh <- msg
}

// Listen forwards outbound messages to target Players
// outside of the interface
func (ge *gameEngine) Listen() {
	for {
		select {
		case joiner := <-ge.registerCh:
			ps := ge.Players()
			ge.players = AddPlayer(ps, joiner)
			for _, p := range ge.players {
				if p.ID() == joiner.ID() {
					continue
				}
				outbound := buildNewJoinerMessage(joiner, p)
				p.Send(outbound)
			}

		case leaver := <-ge.unregisterCh:
			fmt.Println("THIS HAPPENED")
			ps := ge.Players()
			target, ok := ps.Find(leaver.ID())
			if ok {
				underlyingPlayer, typeOK := target.(*WSPlayer)
				if !typeOK {
					panic("this shouldn't have happened")
				}
				underlyingPlayer.conn = nil
			}

		case msg := <-ge.inboundCh:
			switch msg.Command {
			case protocol.Start:
				ge.Start()
				for _, p := range ge.players {
					outbound := buildGameHasStartedMessage(p)
					p.Send(outbound)
				}
			}

			// for all other cases, send to "Game"
			// case protocol.Reorg:

		}
	}
}

func (ge *gameEngine) ID() string {
	return ge.id
}

func (ge *gameEngine) CreatorID() string {
	return ge.creatorID
}

func (ge *gameEngine) Players() Players {
	return ge.players
}

func (ge *gameEngine) PlayState() PlayState {
	return ge.playState
}

func buildGameHasStartedMessage(recipient Player) OutboundMessage {
	return OutboundMessage{
		PlayerID: recipient.ID(),
		Name:     recipient.Name(),
		Message:  fmt.Sprintf("STARTED"),
		Command:  protocol.HasStarted,
	}
}

func buildNewJoinerMessage(joiner, recipient Player) OutboundMessage {
	return OutboundMessage{
		PlayerID: recipient.ID(),
		Name:     recipient.Name(),
		Message:  fmt.Sprintf("%s has joined the game!", joiner.Name()),
		Command:  protocol.NewJoiner,
	}
}

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
