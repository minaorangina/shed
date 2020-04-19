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

func namesToPlayers(names []string) ([]Player, error) {
	players := make([]Player, 0, len(names))
	for _, name := range names {
		p, playerErr := NewPlayer(name)
		if playerErr != nil {
			return []Player{}, playerErr
		}
		players = append(players, p)
	}

	return players, nil
}

// NewGame instantiates a new game of Shed
func NewGame(playerNames []string) *Game {
	cards := deck.New()
	players, _ := namesToPlayers(playerNames)
	return &Game{"Shed", &players, cards}
}

func (g Game) start() {

}
