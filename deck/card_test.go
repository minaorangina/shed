package deck

import (
	"fmt"
	"math/rand"
	"testing"
)

func failureMessage(params ...string) string {
	args := []interface{}{}
	for _, param := range params {
		args = append(args, param)
	}
	return fmt.Sprintf("\nExpected: %s\nActual: %s", args...)
}

func TestCard(t *testing.T) {

	// produces correct string representation of itself
	// zero value card
	zeroValueCard := Card{}
	fmt.Println(zeroValueCard.Rank, zeroValueCard.Suit)
	if zeroValueCard.String() != "Joker" {
		t.Errorf(failureMessage("Joker", zeroValueCard.String()))
	}

	// any card with a rank of zero reliably produces a Joker
	mustBeAJoker := Card{Rank: 0, Suit: Suit(rand.Intn(4))}
	if mustBeAJoker.String() != "Joker" {
		t.Errorf(failureMessage("Joker", mustBeAJoker.String()))
	}

	lowestValueCard := Card{1, 0}
	fmt.Println(lowestValueCard.Rank, lowestValueCard.Suit)
	if lowestValueCard.String() != "Ace of Clubs" {
		t.Errorf(failureMessage("Ace of Clubs", lowestValueCard.String()))
	}

	// specific card
	specificCard := Card{Rank: 12, Suit: 2}
	fmt.Println(specificCard.Rank, specificCard.Suit)
	if specificCard.String() != "Queen of Hearts" {
		t.Errorf(failureMessage("Queen of Hearts", specificCard.String()))
	}
}
