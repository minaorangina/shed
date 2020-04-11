package deck

import (
	"fmt"
	"math/rand"
	"testing"
)

func failureMessage(expected, actual string) string {
	return fmt.Sprintf("\nExpected: %s\nActual: %s", expected, actual)
}
func tableFailureMessage(testName, expected, actual string) string {
	return fmt.Sprintf("%s\nExpected: %s\nActual: %s", testName, expected, actual)
}
func TestRankCard(t *testing.T) {
	// lowest value card
	lowestValueCard := NewRankCard(0, 0)
	if lowestValueCard.String() != "Ace of Clubs" {
		t.Errorf(failureMessage("Ace of Clubs", lowestValueCard.String()))
	}

	// specific card
	specificCard := NewRankCard(11, 2)
	if specificCard.String() != "Queen of Hearts" {
		t.Errorf(failureMessage("Queen of Hearts", specificCard.String()))
	}

	// end of range
	lastCard := NewRankCard(12, 3)
	if lastCard.String() != "King of Spades" {
		t.Errorf(failureMessage("King of Spades", lastCard.String()))
	}

	// out of range (should panic)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected to panic, but it didn't")
			}
		}()
		NewRankCard(13, 2)
		NewRankCard(4, 4)
	}()

	// get rank
	six := NewRankCard(5, rand.Intn(4))
	if six.Rank() != "Six" {
		t.Errorf(failureMessage("Six", six.Rank()))
	}

	// get suit
	spade := NewRankCard(rand.Intn(13), 3)
	if spade.Suit() != "Spades" {
		t.Errorf(failureMessage("Spades", spade.Suit()))
	}
}
