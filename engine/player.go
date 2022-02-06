package engine

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minaorangina/shed/game"
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

// Player represents a player in the game
type Player interface {
	ID() string
	Name() string
	Info() protocol.Player
	Cards() *game.PlayerCards
	Send(msg protocol.OutboundMessage) error
	Receive(data []byte)
}

type WSPlayer struct {
	game.PlayerCards
	id     string
	name   string
	conn   *websocket.Conn // think about how to mock this out
	sendCh chan []byte
	ge     GameEngine
}

// NewPlayer constructs a new player
func NewWSPlayer(id, name string, ws *websocket.Conn, sendCh chan []byte, engine GameEngine) Player {
	player := &WSPlayer{
		id:     id,
		name:   name,
		conn:   ws,
		sendCh: sendCh,
		ge:     engine,
	}

	go player.writePump()
	go player.readPump()

	return player
}

func (p *WSPlayer) Info() protocol.Player {
	return protocol.Player{
		PlayerID: p.id,
		Name:     p.name,
	}
}

func (p *WSPlayer) ID() string {
	return p.id
}

func (p *WSPlayer) Name() string {
	return p.name
}

// Cards returns all of a player's cards
func (p *WSPlayer) Cards() *game.PlayerCards {
	return &game.PlayerCards{
		Hand:   p.Hand,
		Seen:   p.Seen,
		Unseen: p.Unseen,
	}
}

// Send formats a protocol.OutboundMessage and forwards to ws connection
func (p *WSPlayer) Send(msg protocol.OutboundMessage) error {
	var formattedMsg []byte

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	formattedMsg = payload

	// should this be in a goroutine?
	p.sendCh <- formattedMsg

	return nil
}

func (p *WSPlayer) Receive(msg []byte) {
	// convert to InboundMessage

	// put on the game engine chan
}

func (p *WSPlayer) readPump() {
	defer func() {
		p.ge.RemovePlayer(p)
		p.conn.Close()
	}()

	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// switch on message contents
		var inbound protocol.InboundMessage

		err = json.Unmarshal(message, &inbound)
		if err != nil {
			log.Printf("error unmarshalling json: %v", err)
			continue
		}

		log.Printf("lgr (readPump) %s: %+v", time.Now().Format(time.StampMilli), inbound)

		p.ge.Receive(inbound)
	}
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

			log.Printf("lgr (writePump) %s: %+v\n\n", time.Now().Format(time.StampMilli), string(msg))

			err := p.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				panic(err)
				// return
			}

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("removing player %s", p.ID())
				p.ge.RemovePlayer(p)
				// return
			}
		}
	}
}

// Players represents all players in the game
type Players []Player

// NewPlayers returns a set of Players
func NewPlayers(p ...Player) Players {
	return Players(p)
}

// AppendPlayer adds a player to a set of Players
func AppendPlayer(ps Players, p Player) Players {
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

func (ps Players) IDs() []string {
	ids := []string{}
	for _, p := range ps {
		ids = append(ids, p.ID())
	}
	return ids
}

func (ps Players) Info() []protocol.Player {
	info := []protocol.Player{}
	for _, p := range ps {
		info = append(info, p.Info())
	}
	return info
}
