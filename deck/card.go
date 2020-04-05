package deck

import "fmt"

// Rank represents a rank in a deck of cards
type Rank int

var rankNames = []string{"Joker", "Ace", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Jack", "Queen", "King"}

const (
	Joker = iota
	Ace
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

// Suit represents a suit in a deck of cards
type Suit int

var suitNames = []string{"Clubs", "Diamonds", "Hearts", "Spades"}

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

// Card represents a playing card
type Card struct {
	Rank Rank
	Suit Suit
	// maybe a property to represent Jack, King, Queen, Joker etc
}

func (c Card) String() string {
	if c.Rank == Joker {
		return "Joker"
	}
	return fmt.Sprintf("%s of %s", rankNames[c.Rank], suitNames[c.Suit])
}
