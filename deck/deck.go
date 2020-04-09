package deck

import "fmt"

// Deck represents a deck of cards
type Deck []Card

// New creates a deck of cards
func New() Deck {
	// all cards (option to leave out jokers)
	cards := []Card{}
	for suit := range suitNames {
		for rank := range rankNames {
			c := NewRankCard(rank, suit)
			cards = append(cards, c)
		}
	}
	return cards
}

// Shuffle shuffles the deck of cards
func (d *Deck) Shuffle() {

}
