package deck

import (
	"math/rand"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestCard(t *testing.T) {
	cases := []struct {
		name     string
		card     Card
		expected string
	}{
		{"Lowest value card", newCard(0, 0), "Ace of Clubs"},
		{"Specific card", newCard(11, 2), "Queen of Hearts"},
		{"Highest value card", newCard(12, 3), "King of Spades"},
	}

	for _, c := range cases {
		utils.AssertEqual(t, c.card.String(), c.expected)
	}

	t.Run("Out of range (should panic)", func(t *testing.T) {
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
	})

	t.Run("get rank", func(t *testing.T) {
		six := newCard(5, rand.Intn(4))
		utils.AssertEqual(t, six.Rank.String(), "Six")
	})

	t.Run("get suit", func(t *testing.T) {
		spade := newCard(rand.Intn(13), 3)
		utils.AssertEqual(t, spade.Suit.String(), "Spades")
	})
}
