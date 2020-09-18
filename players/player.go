package players

import (
	"io"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
	uuid "github.com/satori/go.uuid"
)

// NewID constructs a player ID
func NewID() string {
	return uuid.NewV4().String()
}

// Conn represents a connection to a player in the real world
type conn struct {
	In  io.Reader
	Out io.Writer
}

// Players represents all players in the game
type Players []*Player

// NewPlayers returns a set of Players
func NewPlayers(p ...*Player) Players {
	return Players(p)
}

// AddPlayer adds a player to a set of Players
func AddPlayer(ps *Players, p *Player) Players {
	if _, ok := ps.Find(p.ID); !ok {
		return Players(append(*ps, p))
	}
	return *ps
}

// Find finds a player by id
func (ps *Players) Find(id string) (*Player, bool) {
	for _, p := range *ps {
		if p.ID == id {
			return p, true
		}
	}
	return nil, false
}

// Player represents a player in the game
type Player struct {
	Hand   []deck.Card
	Seen   []deck.Card
	Unseen []deck.Card
	ID     string
	Name   string
	Conn   *conn // tcp or command line
}

type PlayerCards struct {
	Hand   []deck.Card
	Seen   []deck.Card
	Unseen []deck.Card
}

// NewPlayer constructs a new player
func NewPlayer(id, name string, in io.Reader, out io.Writer) *Player {
	conn := &conn{in, out}
	return &Player{ID: id, Name: name, Conn: conn}
}

// Cards returns all of a player's cards
func (p Player) Cards() PlayerCards {
	return PlayerCards{
		Hand:   p.Hand,
		Seen:   p.Seen,
		Unseen: p.Unseen,
	}
}

func (p Player) SendMessageAwaitReply(ch chan InboundMessage, msg OutboundMessage) {
	switch msg.Command {
	case protocol.Reorg:
		ch <- p.handleReorg(msg)
	}
}

func (p Player) handleReorg(msg OutboundMessage) InboundMessage {
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

	if shouldReorganise := offerCardSwitch(p.Conn, offerTimeout); shouldReorganise {
		response = reorganiseCards(p.Conn, msg)
	}

	return response
}
