package shed

import (
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
	uuid "github.com/satori/go.uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
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

type Conn interface {
	Send(data []byte) error
	Receive(data []byte)
	Conn() io.Closer
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
	conn *websocket.Conn // think about how to mock this out
	send chan []byte
}

// NewPlayer constructs a new player
func NewWSPlayer(id, name string, ws *websocket.Conn) Player {
	player := &WSPlayer{id: id, name: name, conn: ws, send: make(chan []byte)}
	go player.writePump()
	return player
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
	var formattedMsg []byte
	switch msg.Command {
	case protocol.Reorg:

	case protocol.NewJoiner:
		formattedMsg = append([]byte(msg.Message), []byte(" has joined the game!")...)

		// convert to appropriate format

		// call Send on the connection

	}

	p.send <- formattedMsg

	return nil
}

func (p *WSPlayer) Receive(msg []byte) {
	// convert to InboundMessage

	// put on the game engine chan
}

func (p *WSPlayer) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
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
