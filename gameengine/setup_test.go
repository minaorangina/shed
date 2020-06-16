package gameengine

import (
	"fmt"
	"testing"
)

func TestSetupFn(t *testing.T) {
	t.Run("all cards dealt", func(t *testing.T) {
		t.Skip("do not run TestGameEngineStart/all_cards_dealt")
		ge, _ := gameEngineWithPlayers()
		err := ge.Start() // mock required
		if err != nil {
			t.Fatal("Unexpected error ", err.Error())
		}
		for _, p := range ge.Players() {
			c := p.Cards()
			numHand := len(c.Hand)
			numSeen := len(c.Seen)
			numUnseen := len(c.Unseen)
			if numHand != 3 {
				formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
				t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
			}
		}
	})

	t.Run("correct playState", func(t *testing.T) {
		t.Skip("skip testing playstates")
	})
}
