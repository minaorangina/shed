package gameengine

import "github.com/minaorangina/shed/deck"

// Player represents a player in the game
type Player struct {
	Name  string
	cards *playerCards
}

type playerCards struct {
	hand   *[]deck.Card
	seen   *[]deck.Card
	unseen *[]deck.Card
}

// NewPlayer constructs a new player
func NewPlayer(name string) (Player, error) {
	pc := playerCards{}
	return Player{name, &pc}, nil
}
