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
		{"Lowest value card", NewCard(0, 0), "Null of Null"},
		{"Lowest value card", NewCard(1, 1), "Ace of Clubs"},
		{"Specific card", NewCard(12, 3), "Queen of Hearts"},
		{"Highest value card", NewCard(13, 4), "King of Spades"},
	}

	for _, c := range cases {
		utils.AssertEqual(t, c.card.String(), c.expected)
	}

	t.Run("invalid cards (should panic)", func(t *testing.T) {
		utils.ShouldPanic(t, func() { NewCard(14, 2) })
		utils.ShouldPanic(t, func() { NewCard(4, 5) })
		utils.ShouldPanic(t, func() { NewCard(0, 4) })
		utils.ShouldPanic(t, func() { NewCard(4, 0) })
	})

	t.Run("get rank", func(t *testing.T) {
		six := NewCard(Rank(6), Suit(rand.Intn(4)))
		utils.AssertEqual(t, six.Rank.String(), "Six")
	})

	t.Run("get suit", func(t *testing.T) {
		spade := NewCard(Rank(rand.Intn(13)), Suit(4))
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
			{
				NewCard(NullRank, NullSuit),
				WireCard{
					Rank:          "Null",
					Suit:          "Null",
					CanonicalName: "Null of Null",
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
			{
				WireCard{
					Rank:          "Null",
					Suit:          "Null",
					CanonicalName: "Null of Null",
				},
				NewCard(NullRank, NullSuit),
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
		tt := []struct {
			name string
			wc   WireCard
			want Card
		}{
			{
				"Ace of Spades",
				WireCard{"Ace", "Spades", "Ace of Spades"},
				NewCard(Ace, Spades),
			},
			{
				"Null of Null",
				WireCard{"Null", "Null", "Null of Null"},
				NewCard(NullRank, NullSuit),
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				bytes, err := json.Marshal(tc.wc)
				utils.AssertNoError(t, err)

				var got Card
				err = json.Unmarshal(bytes, &got)
				utils.AssertNoError(t, err)
				utils.AssertDeepEqual(t, got, tc.want)
			})
		}
	})

	t.Run("marshal Card to WireCard json", func(t *testing.T) {
		tt := []struct {
			name string
			card Card
			want []byte
		}{
			{
				"Four of Diamonds",
				NewCard(Four, Diamonds),
				[]byte(`{"rank":"Four","suit":"Diamonds","canonicalName":"Four of Diamonds"}`),
			},
			{
				"Null of Null",
				NewCard(NullRank, NullSuit),
				[]byte(`{"rank":"Null","suit":"Null","canonicalName":"Null of Null"}`),
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				got, err := json.Marshal(tc.card)
				utils.AssertNoError(t, err)
				utils.AssertStringEquality(t, string(got), string(tc.want))
			})
		}
	})
}
