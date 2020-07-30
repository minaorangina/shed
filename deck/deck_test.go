package deck

import (
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

var fullDeckCount = 52

func TestNewDeck(t *testing.T) {
	deckOfCards := New()

	if len(deckOfCards) != fullDeckCount {
		utils.FailureMessage(t, len(deckOfCards), fullDeckCount)
	}
}
func TestDeckShuffle(t *testing.T) {
	deckOfCards := New()
	anotherDeckOfCards := New()
	anotherDeckOfCards.Shuffle()
	if reflect.DeepEqual(deckOfCards, anotherDeckOfCards) {
		t.Errorf("Deck was not shuffled")
	}
	if len(deckOfCards) != len(anotherDeckOfCards) {
		utils.FailureMessage(t, len(deckOfCards), len(anotherDeckOfCards))
	}
	func() {
		seenCards := map[string]bool{}
		for i, card := range anotherDeckOfCards {
			if _, val := seenCards[card.String()]; !val {
				seenCards[card.String()] = true
			} else {
				t.Fatalf("Duplicate card found at index %d: %s", i, card.String())
			}
		}
	}()
}
func TestDeckRemoveCards(t *testing.T) {
	deckToRemoveFrom := New()
	type deckTest struct {
		testName             string
		numCardsToRemove     int
		expectedNumRemaining int
	}

	zeroCards := deckToRemoveFrom.Deal(0)
	if len(zeroCards) != 0 {
		utils.FailureMessage(t, len(zeroCards), 0)
	}

	zeroCards = deckToRemoveFrom.Deal(-5)
	if len(zeroCards) != 0 {
		utils.FailureMessage(t, len(zeroCards), 0)
	}

	deckTests := []deckTest{
		{"remove first card", 1, 51},
		{"remove second card", 1, 50},
		{"remove four cards", 4, 46},
		{"remove all but one card", 45, 1},
		{"remove last card", 1, 0},
	}

	for _, dt := range deckTests {
		expectedCardName := deckToRemoveFrom[len(deckToRemoveFrom)-1].String()
		removedCards := deckToRemoveFrom.Deal(dt.numCardsToRemove)
		if len(removedCards) != dt.numCardsToRemove {
			utils.TableFailureMessage(t, dt.testName, len(removedCards), dt.numCardsToRemove)
		}
		if removedCards[dt.numCardsToRemove-1].String() != expectedCardName {
			utils.TableFailureMessage(t, dt.testName, removedCards[dt.numCardsToRemove-1].String(), expectedCardName)
		}
		if len(deckToRemoveFrom) != dt.expectedNumRemaining {
			utils.TableFailureMessage(t, dt.testName, len(deckToRemoveFrom), dt.expectedNumRemaining)
		}
	}
}
