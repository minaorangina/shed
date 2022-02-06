package engine

import (
	"errors"
	"log"
	"sync"
	"time"

	gm "github.com/minaorangina/shed/game"
	"github.com/minaorangina/shed/protocol"
)

var ErrNilGame = errors.New("game is nil")

// PlayState represents the state of the current game
// idle -> no game play (pre game and post game)
// InProgress -> game in progress
// paused -> game is paused
type PlayState int

func (gps PlayState) String() string {
	if gps == 0 {
		return "Idle"
	} else if gps == 1 {
		return "InProgress"
	} else if gps == 2 {
		return "Paused"
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
	Send([]protocol.OutboundMessage)
	Players() Players
	ID() string
	CreatorID() string
	AddPlayer(Player) error
	RemovePlayer(Player)
	Receive(protocol.InboundMessage)
	PlayState() PlayState
	Game() gm.Game
}

// gameEngine represents the engine of the game

type gameEngine struct {
	id                       string
	creatorID                string
	playState                PlayState
	players                  Players
	registerCh, unregisterCh chan Player
	inboundCh                chan protocol.InboundMessage
	outboundCh               chan []protocol.OutboundMessage
	gameCh                   chan []protocol.InboundMessage
	game                     gm.Game
}

// GameEngineOpts represents options for constructing a new GameEngine
type GameEngineOpts struct {
	GameID                   string
	CreatorID                string
	Players                  Players
	RegisterCh, UnregisterCh chan Player
	InboundCh                chan protocol.InboundMessage
	OutboundCh               chan []protocol.OutboundMessage
	GameCh                   chan []protocol.InboundMessage
	PlayState                PlayState
	Game                     gm.Game
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
		opts.InboundCh = make(chan protocol.InboundMessage)
	}
	if opts.OutboundCh == nil {
		opts.OutboundCh = make(chan []protocol.OutboundMessage)
	}
	if opts.GameCh == nil {
		opts.GameCh = make(chan []protocol.InboundMessage)
	}
	engine := &gameEngine{
		id:           opts.GameID,
		creatorID:    opts.CreatorID,
		players:      opts.Players,
		registerCh:   opts.RegisterCh,
		unregisterCh: opts.UnregisterCh,
		inboundCh:    opts.InboundCh,
		outboundCh:   opts.OutboundCh,
		gameCh:       opts.GameCh,
		playState:    opts.PlayState,
		game:         opts.Game,
	}

	// Listen for websocket connections
	go engine.Listen()

	return engine, nil
}

// Start starts a game
func (ge *gameEngine) Start() error {
	if ge.playState != Idle {
		return nil
	}
	if ge.game == nil {
		return ErrNilGame
	}
	err := ge.game.Start(ge.players.Info())
	if err != nil {
		return err
	}

	// mutex
	ge.playState = InProgress
	go ge.Play()

	return nil
}

func (ge *gameEngine) Play() {
	// this will need some way to shutdown gracefully
	for inbound := range ge.gameCh {
		var (
			outbound []protocol.OutboundMessage
			err      error
		)

		if len(inbound) == 0 {
			outbound, err = ge.game.Next()
		} else {
			outbound, err = ge.game.ReceiveResponse(inbound)
		}
		if err != nil {
			log.Printf("error: %s\n%v", err.Error(), outbound)
		}
		// what happens if outbound == nil?
		ge.Send(outbound)
	}
}

// Listen forwards outbound messages to target Players
// outside of the interface
func (ge *gameEngine) Listen() {
	commTracker := struct {
		mu              *sync.Mutex
		messages        []protocol.InboundMessage
		expectedCommand protocol.Cmd
	}{
		mu:              &sync.Mutex{},
		messages:        []protocol.InboundMessage{},
		expectedCommand: protocol.Reorg,
	}

	for {
		select {
		case joiner := <-ge.registerCh:
			ps := ge.Players()
			ge.players = AppendPlayer(ps, joiner)
			for _, p := range ge.players {
				if p.ID() == joiner.ID() {
					continue
				}
				outbound := gm.BuildNewJoinerMessage(p.ID(), p.Name(), joiner.ID(), joiner.Name())
				p.Send(outbound)
			}

		case leaver := <-ge.unregisterCh:
			ps := ge.Players()
			target, ok := ps.Find(leaver.ID())
			if ok {
				underlyingPlayer, typeOK := target.(*WSPlayer)
				if !typeOK {
					panic("this shouldn't have happened")
				}
				underlyingPlayer.conn = nil
			}

		case msgs := <-ge.outboundCh:
			ge.messagePlayers(msgs)
			if !ge.game.GameOver() && ge.game.AwaitingResponse() == protocol.Null {
				ge.sendToGame(nil)
			}

		case msg := <-ge.inboundCh:
			if msg.Command == protocol.Start {

				if err := ge.Start(); err != nil {
					p, _ := ge.players.Find(msg.PlayerID)

					p.Send(protocol.OutboundMessage{
						PlayerID: msg.PlayerID,
						Command:  protocol.Error,
						Error:    err.Error(),
					})
					continue
				}

				for _, p := range ge.players {
					p.Send(gm.BuildGameHasStartedMessage(p.ID(), p.Name()))
				}
				// small delay before game starts
				<-time.After(time.Millisecond * 400)
				ge.sendToGame(nil)

				continue
			}

			// Ignore messages that are not expected
			if msg.Command != ge.game.AwaitingResponse() {
				log.Printf("lgr: unexpected cmd %s, ignoring\n", msg.Command)
				continue
			}

			switch msg.Command {
			case protocol.Reorg:
				commTracker.mu.Lock()
				commTracker.messages = append(commTracker.messages, msg)
				// send back
				commTracker.mu.Unlock()

				if len(commTracker.messages) == len(ge.Players()) {
					commTracker.mu.Lock()
					log.Printf("lgr %s: all players have reorg'd", time.Now().Format(time.StampMilli))
					ge.sendToGame(commTracker.messages)

					commTracker.messages = []protocol.InboundMessage{}
					commTracker.expectedCommand = protocol.Null

					commTracker.mu.Unlock()
				}

			default:
				ge.sendToGame([]protocol.InboundMessage{msg}) // handle failures
			}
		}
	}
}

func (ge *gameEngine) messagePlayers(msgs []protocol.OutboundMessage) {
	for _, m := range msgs {
		p, ok := ge.players.Find(m.PlayerID)
		if ok {
			p.Send(m)
		}
	}
}

func (ge *gameEngine) sendToGame(msgs []protocol.InboundMessage) {
	ge.gameCh <- msgs
}

func (ge *gameEngine) Send(msgs []protocol.OutboundMessage) {
	ge.outboundCh <- msgs
}

// Receive forwards protocol.InboundMessages from Players for sorting
func (ge *gameEngine) Receive(msg protocol.InboundMessage) {
	ge.inboundCh <- msg
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

func (ge *gameEngine) Game() gm.Game {
	return ge.game
}
