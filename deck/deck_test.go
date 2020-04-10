package deck

import (
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

	// Take a card off the deck
	deckToRemoveFrom := New()

	// Empty list of cards
	zeroCards := deckToRemoveFrom.Deal(0)
	if len(zeroCards) != 0 {
		t.Errorf(failureMessage(strconv.Itoa(0), strconv.Itoa(len(zeroCards))))
	}

	numCardsToRemove := 1
	expectedNumRemaining := 51
	expectedCardName := deckToRemoveFrom[expectedNumRemaining].String()
	oneCard := deckToRemoveFrom.Deal(numCardsToRemove)
	if len(oneCard) != numCardsToRemove {
		t.Errorf(failureMessage(strconv.Itoa(numCardsToRemove), strconv.Itoa(len(oneCard))))
	}
	if oneCard[0].String() != expectedCardName {
		t.Errorf(failureMessage(expectedCardName, oneCard[0].String()))
	}
	if len(deckToRemoveFrom) != expectedNumRemaining {
		t.Errorf(failureMessage(strconv.Itoa(expectedNumRemaining), strconv.Itoa(len(deckToRemoveFrom))))
	}

	// Take off another card from the same deck
	numCardsToRemove = 1
	expectedNumRemaining = 50
	expectedCardName = deckToRemoveFrom[expectedNumRemaining].String()
	oneCard = deckToRemoveFrom.Deal(numCardsToRemove)
	if len(oneCard) != numCardsToRemove {
		t.Errorf(failureMessage(strconv.Itoa(numCardsToRemove), strconv.Itoa(len(oneCard))))
	}
	if oneCard[0].String() != expectedCardName {
		t.Errorf(failureMessage(expectedCardName, oneCard[0].String()))
	}
	if len(deckToRemoveFrom) != expectedNumRemaining {
		t.Errorf(failureMessage(strconv.Itoa(expectedNumRemaining), strconv.Itoa(len(deckToRemoveFrom))))
	}

	// Take off 50 cards
	numCardsToRemove = 49
	expectedNumRemaining = 1
	expectedCardName = deckToRemoveFrom[expectedNumRemaining].String()
	removedCards := deckToRemoveFrom.Deal(numCardsToRemove)
	if len(removedCards) != numCardsToRemove {
		t.Errorf(failureMessage(strconv.Itoa(numCardsToRemove), strconv.Itoa(len(removedCards))))
	}
	if len(deckToRemoveFrom) != expectedNumRemaining {
		t.Errorf(failureMessage(strconv.Itoa(expectedNumRemaining), strconv.Itoa(len(deckToRemoveFrom))))
	}

	// Take off last card
	numCardsToRemove = 1
	expectedNumRemaining = 0
	expectedCardName = deckToRemoveFrom[expectedNumRemaining].String()
	removedCards = deckToRemoveFrom.Deal(numCardsToRemove)
	if len(removedCards) != numCardsToRemove {
		t.Errorf(failureMessage(strconv.Itoa(numCardsToRemove), strconv.Itoa(len(removedCards))))
	}
	if len(deckToRemoveFrom) != expectedNumRemaining {
		t.Errorf(failureMessage(strconv.Itoa(expectedNumRemaining), strconv.Itoa(len(deckToRemoveFrom))))
	}

	// Deal on an empty deck returns an empty slice of cards
	empty := deckToRemoveFrom.Deal(1)
	if len(empty) != 0 {
		t.Errorf(failureMessage(strconv.Itoa(0), strconv.Itoa(len(empty))))
	}
}
