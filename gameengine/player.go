package gameengine

import "github.com/minaorangina/shed/deck"

// Player represents a player in the game
type Player struct {
	id    int
	name  string
	cards *playerCards
}

type playerCards struct {
	hand   *[]deck.Card
	seen   *[]deck.Card
	unseen *[]deck.Card
}

// NewPlayer constructs a new player
func NewPlayer(id int, name string) (Player, error) {
	pc := playerCards{}
	return Player{id: id, name: name, cards: &pc}, nil
}
