package gameengine

import (
	"github.com/minaorangina/shed/deck"
	uuid "github.com/satori/go.uuid"
)

// NewID constructs a player ID
func NewID() string {
	return uuid.NewV4().String()
}

// Player represents a player in the game
type Player struct {
	id     string
	name   string
	hand   []deck.Card
	seen   []deck.Card
	unseen []deck.Card
}

type playerCards struct {
	hand   []deck.Card
	seen   []deck.Card
	unseen []deck.Card
}

// NewPlayer constructs a new player
func NewPlayer(id, name string) Player {
	return Player{id: id, name: name}
}

func (p Player) cards() playerCards {
	return playerCards{
		hand:   p.hand,
		seen:   p.seen,
		unseen: p.unseen,
	}
}
