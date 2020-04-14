package gameengine

import (
	"github.com/minaorangina/shed/deck"
)

// Game represents a game
type Game struct {
	Name    string
	players *[]Player
	deck    deck.Deck
}

// Stage represents the main stages in the game
type Stage int

const (
	handOrganisation Stage = iota
	clearDeck
	clearHand
)

// NewGame instantiates a new game of Shed
func NewGame(players *[]Player) Game {
	cards := deck.New()
	return Game{"Shed", players, cards}
}

func (g Game) start() {

}
