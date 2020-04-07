package deck

import (
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
}
