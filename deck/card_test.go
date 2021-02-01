package deck

import (
	"encoding/json"
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

func TestWireCard(t *testing.T) {
	t.Run("type conversion from Card", func(t *testing.T) {
		tt := []struct {
			source Card
			want   WireCard
		}{
			{
				NewCard(Ace, Spades),
				WireCard{
					Rank:          "Ace",
					Suit:          "Spades",
					CanonicalName: "Ace of Spades",
				},
			},
		}

		for _, tc := range tt {
			t.Run(tc.source.String(), func(t *testing.T) {
				got := tc.source.ToWireCard()
				utils.AssertDeepEqual(t, got, tc.want)
			})
		}
	})

	t.Run("type conversion to Card", func(t *testing.T) {
		tt := []struct {
			source WireCard
			want   Card
		}{
			{
				WireCard{
					Rank:          "Ace",
					Suit:          "Spades",
					CanonicalName: "Ace of Spades",
				},
				NewCard(Ace, Spades),
			},
		}

		for _, tc := range tt {
			t.Run(tc.source.String(), func(t *testing.T) {
				got := tc.source.ToCard()
				utils.AssertDeepEqual(t, got, tc.want)
			})
		}
	})

	t.Run("unmarshal WireCard json to Card", func(t *testing.T) {
		wc := WireCard{"Ace", "Spades", "Ace of Spades"}
		bytes, err := json.Marshal(wc)
		utils.AssertNoError(t, err)

		want := NewCard(Ace, Spades)

		var got Card
		err = json.Unmarshal(bytes, &got)
		utils.AssertNoError(t, err)
		utils.AssertDeepEqual(t, got, want)
	})

	t.Run("marshal Card to WireCard json", func(t *testing.T) {
		want := []byte(`{"rank":"Four","suit":"Diamonds","canonicalName":"Four of Diamonds"}`)
		card := NewCard(Four, Diamonds)

		got, err := json.Marshal(card)
		utils.AssertNoError(t, err)
		utils.AssertStringEquality(t, string(got), string(want))
	})
}
