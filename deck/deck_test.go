package deck

import (
	"reflect"
	"strconv"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

var fullDeckCount = 52

func TestNewDeck(t *testing.T) {
	deckOfCards := New()

	if len(deckOfCards) != fullDeckCount {
		t.Errorf(utils.FailureMessage(strconv.Itoa(fullDeckCount), strconv.Itoa(len(deckOfCards))))
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
		t.Errorf(utils.FailureMessage(strconv.Itoa(len(deckOfCards)), strconv.Itoa(len(anotherDeckOfCards))))
	}
	func() {
		seenCards := map[string]bool{}
		for i, card := range anotherDeckOfCards {
			if _, val := seenCards[card.String()]; !val {
				seenCards[card.String()] = true
			} else {
				t.Errorf("Duplicate card found at index %d: %s", i, card.String())
				break
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
		t.Errorf(utils.FailureMessage(0, len(zeroCards)))
	}

	zeroCards = deckToRemoveFrom.Deal(-5)
	if len(zeroCards) != 0 {
		t.Errorf(utils.FailureMessage(0, len(zeroCards)))
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
			t.Errorf(utils.TableFailureMessage(dt.testName, strconv.Itoa(dt.numCardsToRemove), strconv.Itoa(len(removedCards))))
		}
		if removedCards[dt.numCardsToRemove-1].String() != expectedCardName {
			t.Errorf(utils.TableFailureMessage(dt.testName, expectedCardName, removedCards[dt.numCardsToRemove-1].String()))
		}
		if len(deckToRemoveFrom) != dt.expectedNumRemaining {
			t.Errorf(utils.TableFailureMessage(dt.testName, strconv.Itoa(dt.expectedNumRemaining), strconv.Itoa(len(deckToRemoveFrom))))
		}
	}
}
