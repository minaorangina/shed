package deck

import (
	"fmt"
)

// Rank represents a rank in a deck of cards
type Rank int

var rankNames = []string{"Ace", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Jack", "Queen", "King"}

const (
	Ace Rank = iota
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

func (r Rank) String() string {
	return rankNames[r]
}

// Suit represents a suit in a deck of cards
type Suit int

var suitNames = []string{"Clubs", "Diamonds", "Hearts", "Spades"}

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

func (s Suit) String() string {
	return suitNames[s]
}

// Card represents cards that belong to a suit
type Card struct {
	Rank Rank
	Suit Suit
	// maybe a property to represent Jack, King, Queen, Joker etc
}

// NewCard constructs a card
func newCard(rank, suit int) Card {
	if rank < 0 || rank > int(King) || suit < 0 || suit > int(Spades) {
		panic("arguments out of range")
	}
	return Card{Rank: Rank(rank), Suit: Suit(suit)}
}

func (rc Card) String() string {
	return fmt.Sprintf("%s of %s", rankNames[rc.Rank], suitNames[rc.Suit])
}
