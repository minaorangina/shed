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
		{"Lowest value card", NewCard(0, 0), "Ace of Clubs"},
		{"Specific card", NewCard(11, 2), "Queen of Hearts"},
		{"Highest value card", NewCard(12, 3), "King of Spades"},
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
			NewCard(13, 2)
			NewCard(4, 4)
		}()
	})

	t.Run("get rank", func(t *testing.T) {
		six := NewCard(Rank(5), Suit(rand.Intn(4)))
		utils.AssertEqual(t, six.Rank.String(), "Six")
	})

	t.Run("get suit", func(t *testing.T) {
		spade := NewCard(Rank(rand.Intn(13)), Suit(3))
		utils.AssertEqual(t, spade.Suit.String(), "Spades")
	})
}
