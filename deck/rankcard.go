package deck

import (
	"errors"
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

// Suit represents a suit in a deck of cards
type Suit int

var suitNames = []string{"Clubs", "Diamonds", "Hearts", "Spades"}

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

// RankCard represents cards that belong to a suit
type RankCard struct {
	rank Rank
	suit Suit
	// maybe a property to represent Jack, King, Queen, Joker etc
}

// NewRankCard constructs a card
func NewRankCard(rank, suit int) (Card, error) {
	if rank < 0 || rank > int(King) || suit < 0 || suit > int(Spades) {
		return RankCard{}, errors.New("arguments out of range")
	}
	return RankCard{rank: Rank(rank), suit: Suit(suit)}, nil
}

// Rank returns a card's rank
func (rc RankCard) Rank() string {
	return rankNames[rc.rank]
}

// Suit returns a card's suit
func (rc RankCard) Suit() string {
	return suitNames[rc.suit]
}

func (rc RankCard) String() string {
	return fmt.Sprintf("%s of %s", rankNames[rc.rank], suitNames[rc.suit])
}
