package gameengine

import "github.com/minaorangina/shed/deck"

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

type playerInfo struct {
	id   string
	name string
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
