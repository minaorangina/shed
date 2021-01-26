package deck

import (
	"encoding/json"
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
func NewCard(rank Rank, suit Suit) Card {
	if rank < Rank(0) || rank > King || suit < Suit(0) || suit > Spades {
		panic("arguments out of range")
	}
	return Card{Rank: Rank(rank), Suit: Suit(suit)}
}

func (c Card) String() string {
	return fmt.Sprintf("%s of %s", rankNames[c.Rank], suitNames[c.Suit])
}

func (c Card) ToWireCard() WireCard {
	return WireCard{
		Rank: rankNames[c.Rank],
		Suit: suitNames[c.Suit],
	}
}

func (c Card) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToWireCard())
}

func (c *Card) UnmarshalJSON(b []byte) error {
	var wc WireCard
	if err := json.Unmarshal(b, &wc); err != nil {
		return err
	}

	*c = wc.ToCard()
	return nil
}

type WireCard struct {
	Rank string `json:"rank"`
	Suit string `json:"suit"`
}

func (wc WireCard) String() string {
	return fmt.Sprintf("%q of %q", wc.Rank, wc.Suit)
}

func (wc WireCard) ToCard() Card {
	var rank Rank
	var suit Suit

	for i, name := range rankNames {
		if name == wc.Rank {
			rank = Rank(i)
			break
		}
	}

	for i, name := range suitNames {
		if name == wc.Suit {
			suit = Suit(i)
			break
		}
	}

	return NewCard(rank, suit)
}
