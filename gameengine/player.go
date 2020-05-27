package gameengine

import (
	"fmt"
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
func NewPlayer(id, name string, in, out *os.File) Player {
	conn := &conn{in, out}
	return Player{ID: id, Name: name, Conn: conn}
}

func (p Player) cards() playerCards {
	return playerCards{
		hand:   p.hand,
		seen:   p.seen,
		unseen: p.unseen,
	}
}

func (p Player) sendMessageAwaitReply(ch chan messageFromPlayer, msg messageToPlayer) {
	switch msg.Command {
	case reorg:
		ch <- p.handleReorg(msg)
	}
}

func (p Player) handleReorg(msg messageToPlayer) messageFromPlayer {
	response := messageFromPlayer{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Hand:     msg.Hand,
		Seen:     msg.Seen,
	}

	playerCards := playerCards{
		seen: msg.Seen,
		hand: msg.Hand,
	}
	fmt.Printf("%s, here are your cards:\n\n", msg.Name)
	fmt.Println(buildCardDisplayText(playerCards))

	if shouldReorganise := offerCardSwitch(p.Conn); shouldReorganise {
		response = reorganiseCards(p.Conn, msg)
	}

	return response
}

// Players represents all players in the game
type AllPlayers map[string]*Player

// NewAllPlayers constructs a Players object
func NewAllPlayers(players ...Player) AllPlayers {
	all := map[string]*Player{}
	for i := range players {
		all[players[i].ID] = &players[i]
	}

	return AllPlayers(all)
}

// OutboundMessages represents messages destined for players
type OutboundMessages map[string]messageToPlayer

// Add adds a message
func (om OutboundMessages) Add(id string, msg messageToPlayer) {
	om[id] = msg
}

// InboundMessages represents messages from players
type InboundMessages map[string]messageFromPlayer

// Add adds a message
func (im InboundMessages) Add(id string, msg messageFromPlayer) {
	im[id] = msg
}
