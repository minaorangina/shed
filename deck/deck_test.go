package deck

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

var fullDeckCount = 52

func TestDeck(t *testing.T) {
	// New
	deckOfCards := New()

	if len(deckOfCards) != fullDeckCount {
		t.Errorf(failureMessage(strconv.Itoa(fullDeckCount), strconv.Itoa(len(deckOfCards))))
	}

	// Shuffle
	anotherDeckOfCards := New()
	anotherDeckOfCards.Shuffle()
	fmt.Println(deckOfCards, "\n", anotherDeckOfCards)
	if reflect.DeepEqual(deckOfCards, anotherDeckOfCards) {
		t.Errorf("Deck was not shuffled")
	}
	if len(deckOfCards) != len(anotherDeckOfCards) {
		t.Errorf(failureMessage(strconv.Itoa(len(deckOfCards)), strconv.Itoa(len(anotherDeckOfCards))))
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
