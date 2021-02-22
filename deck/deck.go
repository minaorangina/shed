package deck

import (
	"math/rand"
	"time"
)

// Deck represents a deck of cards
type Deck []Card

// New creates a deck of cards
func New() Deck {
	// all cards (option to leave out jokers)
	cards := []Card{}
	for suit := range suitNames[1:] {
		for rank := range rankNames[1:] {
			c := NewCard(Rank(rank), Suit(suit))
			cards = append(cards, c)
		}
	}
	return cards
}

// Shuffle shuffles the deck of cards
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	actualDeck := (*d)
	for i := len(actualDeck) - 1; i > 0; i-- {
		randomNumber := rand.Intn(i)
		actualDeck[i], actualDeck[randomNumber] = actualDeck[randomNumber], actualDeck[i]
	}
}

// Deal deals n number of cards from the deck, until it is empty
// If more cards than exist are requested, Deal returns all the remaining cards
func (d *Deck) Deal(n int) []Card {
	numCardsInDeck := len(*d)
	if n < 0 {
		return []Card{}
	}
	if n > numCardsInDeck {
		n = numCardsInDeck
	}

	startingIndex := numCardsInDeck - n
	subSlice := (*d)[startingIndex:numCardsInDeck]
	*d = (*d)[:startingIndex]
	return subSlice
}
