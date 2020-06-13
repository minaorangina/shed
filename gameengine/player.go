package gameengine

import (
	"io"
	"os"

	"github.com/minaorangina/shed/deck"
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

func (ps Players) Individual(id string) *Player {
	for _, p := range ps {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Player represents a player in the game
type Player struct {
	hand   []deck.Card
	seen   []deck.Card
	unseen []deck.Card
	ID     string
	Name   string
	Conn   *conn // tcp or command line
}

type playerCards struct {
	hand   []deck.Card
	seen   []deck.Card
	unseen []deck.Card
}

// NewPlayer constructs a new player
func NewPlayer(id, name string, in, out *os.File) *Player {
	conn := &conn{in, out}
	return &Player{ID: id, Name: name, Conn: conn}
}

func (p Player) cards() playerCards {
	return playerCards{
		hand:   p.hand,
		seen:   p.seen,
		unseen: p.unseen,
	}
}

func (p Player) sendMessageAwaitReply(ch chan InboundMessage, msg OutboundMessage) {
	switch msg.Command {
	case reorg:
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

	playerCards := playerCards{
		seen: msg.Seen,
		hand: msg.Hand,
	}
	SendText(p.Conn.Out, "%s, here are your cards:\n\n", msg.Name)
	SendText(p.Conn.Out, buildCardDisplayText(playerCards))

	if shouldReorganise := offerCardSwitch(p.Conn); shouldReorganise {
		response = reorganiseCards(p.Conn, msg)
	}

	return response
}
