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
	id     string
	name   string
	conn   *websocket.Conn // think about how to mock this out
	sendCh chan []byte
	engine GameEngine
}

// NewPlayer constructs a new player
func NewWSPlayer(id, name string, ws *websocket.Conn, sendCh chan []byte, engine GameEngine) Player {
	player := &WSPlayer{
		id:     id,
		name:   name,
		conn:   ws,
		sendCh: sendCh,
		engine: engine,
	}
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

// Send formats a OutboundMessage and forwards to ws connection
func (p *WSPlayer) Send(msg OutboundMessage) error {
	var formattedMsg []byte

	switch msg.Command {
	case protocol.Reorg:

	case protocol.NewJoiner:
		formattedMsg = []byte(msg.Message)
	}

	// should this be in a goroutine?
	p.sendCh <- formattedMsg

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
		case msg, ok := <-p.sendCh:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte("Something went wrong"))
			}

			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				panic(err)
				// return
			}
			w.Write(msg)

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				panic(err)
				// return
			}
		}
	}
}
