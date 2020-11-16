package players

import (
	"io"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
	uuid "github.com/satori/go.uuid"
)

// NewID constructs a player ID
func NewID() string {
	return uuid.NewV4().String()
}

// Conn represents a connection to a player in the real world
// type Conn interface {
// 	Send(data []byte) error
// 	Receive(data []byte)
// }

type conn struct {
	In  io.Reader
	Out io.Writer
}

type WSConn struct {
	Out *websocket.Conn
}

func NewWSConn(c *websocket.Conn) *WSConn {
	return &WSConn{c}
}

func (c *WSConn) Send(data []byte) error {
	return nil
}

func (c *WSConn) Receive(data []byte) {

}

type PlayerCards struct {
	Hand   []deck.Card
	Seen   []deck.Card
	Unseen []deck.Card
}

// Player represents a player in the game
type Player interface {
	ID() string
	Name() string
	Cards() *PlayerCards
	Send(msg OutboundMessage) error
	Receive(data []byte)
}

type WSPlayer struct {
	PlayerCards
	id   string
	name string
	conn *WSConn
}

// NewPlayer constructs a new player
func NewWSPlayer(id, name string, ws *websocket.Conn) Player {
	return &WSPlayer{id: id, name: name, conn: &WSConn{ws}}
}

func (p *WSPlayer) ID() string {
	return p.id
}

func (p *WSPlayer) Name() string {
	return p.name
}

// Cards returns all of a player's cards
func (p *WSPlayer) Cards() *PlayerCards {
	return &PlayerCards{
		Hand:   p.Hand,
		Seen:   p.Seen,
		Unseen: p.Unseen,
	}
}

func (p *WSPlayer) Send(msg OutboundMessage) error {
	// check command
	switch msg.Command {
	case protocol.Reorg:

		// convert to appropriate format

		// call Send on the connection

		return nil
	}
	return nil
}

func (p *WSPlayer) Receive(msg []byte) {
	// convert to InboundMessage

	// put on the game engine chan
}

type CLIPlayer struct {
	PlayerCards
	id   string
	name string
	Conn *conn
}

func (p CLIPlayer) ID() string {
	return p.id
}

func (p CLIPlayer) Name() string {
	return p.name
}

// Cards returns all of a player's cards
func (p CLIPlayer) Cards() *PlayerCards {
	return &PlayerCards{
		Hand:   p.Hand,
		Seen:   p.Seen,
		Unseen: p.Unseen,
	}
}

func (p CLIPlayer) Send(msg OutboundMessage) error {
	// check command
	switch msg.Command {
	case protocol.Reorg:
		inbound := p.handleReorg(msg)
		_ = inbound

		// convert to appropriate format

		// call Send on the connection

		return nil
	}
	return nil
}

func (p CLIPlayer) Receive(msg []byte) {
	// convert to InboundMessage

	// put on the game engine chan
}

func (p CLIPlayer) handleReorg(msg OutboundMessage) InboundMessage {
	response := InboundMessage{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Hand:     msg.Hand,
		Seen:     msg.Seen,
	}

	playerCards := PlayerCards{
		Seen: msg.Seen,
		Hand: msg.Hand,
	}
	SendText(p.Conn.Out, "%s, here are your cards:\n\n", msg.Name)
	SendText(p.Conn.Out, buildCardDisplayText(playerCards))

	// could probably make this one step. user submits or skips - no need to
	if shouldReorganise := offerCardSwitch(p.Conn, offerTimeout); shouldReorganise {
		response = reorganiseCards(p.Conn, msg)
	}

	return response
}

// Players represents all players in the game
type Players []Player

// NewPlayers returns a set of Players
func NewPlayers(p ...Player) Players {
	return Players(p)
}

// AddPlayer adds a player to a set of Players
func AddPlayer(ps Players, p Player) Players {
	if _, ok := ps.Find(p.ID()); !ok {
		return Players(append(ps, p))
	}
	return ps
}

// Find finds a player by id
func (ps Players) Find(id string) (Player, bool) {
	for _, p := range ps {
		if got := p.ID(); got == id {
			return p, true
		}
	}
	return nil, false
}
