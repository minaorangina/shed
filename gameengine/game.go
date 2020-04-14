package gameengine

import "github.com/minaorangina/shed/player"

// Game represents a game
type Game struct {
	Name    string
	players *[]player.Player
}

// Stage represents the main stages in the game
type Stage int

const (
	handOrganisation Stage = iota
	clearDeck
	clearHand
)

// NewGame instantiates a new game of Shed
func NewGame(players *[]player.Player) Game {
	return Game{"Shed", players}
}

func (g Game) start() {

}
