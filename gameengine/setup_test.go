package gameengine

import (
	"testing"

	"github.com/minaorangina/shed/deck"
)

func TestHandleInitialCards(t *testing.T) {
	// test a well-formed inbound message
	// test a malformed inbound message

	t.Run("dealUnseenCards", func(t *testing.T) {
		cards := deck.New()
		ps := somePlayers()
		dealUnseenCards(cards, ps)

		for _, p := range ps {
			if len(p.Unseen) != 3 {
				t.Fatalf("Wanted 3, got %d", len(p.Unseen))
			}
		}
	})

	t.Run("dealInitialCards", func(t *testing.T) {
		cards := deck.New()
		ps := somePlayers()
		got := dealInitialCards(cards, ps)

		for _, p := range got {
			if len(p.Seen) != 3 || len(p.Hand) != 3 {
				t.Errorf("dealInitialCards not working as expected, got %+v", got)
			}
		}
	})

	t.Run("confirmInitialCards", func(t *testing.T) {
	})
}
