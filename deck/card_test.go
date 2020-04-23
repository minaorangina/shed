package deck

import (
	"math/rand"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestCard(t *testing.T) {
	// lowest value card
	lowestValueCard := newCard(0, 0)
	if lowestValueCard.String() != "Ace of Clubs" {
		t.Errorf(utils.FailureMessage("Ace of Clubs", lowestValueCard.String()))
	}

	// specific card
	specificCard := newCard(11, 2)
	if specificCard.String() != "Queen of Hearts" {
		t.Errorf(utils.FailureMessage("Queen of Hearts", specificCard.String()))
	}

	// end of range
	lastCard := newCard(12, 3)
	if lastCard.String() != "King of Spades" {
		t.Errorf(utils.FailureMessage("King of Spades", lastCard.String()))
	}

	// out of range (should panic)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected to panic, but it didn't")
			}
		}()
		newCard(13, 2)
		newCard(4, 4)
	}()

	// get rank
	six := newCard(5, rand.Intn(4))
	if six.Rank.String() != "Six" {
		t.Errorf(utils.FailureMessage("Six", six.Rank.String()))
	}

	// get suit
	spade := newCard(rand.Intn(13), 3)
	if spade.Suit.String() != "Spades" {
		t.Errorf(utils.FailureMessage("Spades", spade.Suit.String()))
	}
}
