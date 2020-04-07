package deck

import (
	"fmt"
	"math/rand"
	"testing"
)

func failureMessage(expected, actual string) string {
	return fmt.Sprintf("\nExpected: %s\nActual: %s", expected, actual)
}
func TestRankCard(t *testing.T) {
	// lowest value card
	lowestValueCard, err := NewRankCard(0, 0)
	if err != nil {
		t.Errorf(err.Error())
	}
	if lowestValueCard.String() != "Ace of Clubs" {
		t.Errorf(failureMessage("Ace of Clubs", lowestValueCard.String()))
	}

	// specific card
	specificCard, err := NewRankCard(11, 2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if specificCard.String() != "Queen of Hearts" {
		t.Errorf(failureMessage("Queen of Hearts", specificCard.String()))
	}

	// end of range
	lastCard, err := NewRankCard(12, 3)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if lastCard.String() != "King of Spades" {
		t.Errorf(failureMessage("King of Spades", lastCard.String()))
	}

	// out of range
	_, err = NewRankCard(13, 2)
	if err == nil {
		t.Errorf("Expected error...")
	}

	_, err = NewRankCard(4, 4)
	if err == nil {
		t.Errorf("Expected error...")
	}

	// get rank
	six, err := NewRankCard(5, rand.Intn(4))
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else if six.Rank() != "Six" {
		t.Errorf(failureMessage("Six", six.Rank()))
	}

	// get suit
	spade, err := NewRankCard(rand.Intn(13), 3)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if spade.Suit() != "Spades" {
		t.Errorf(failureMessage("Spades", spade.Suit()))
	}
}
