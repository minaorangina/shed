package shed

import (
	"testing"

	"github.com/minaorangina/shed/deck"
	utils "github.com/minaorangina/shed/internal"
)

func TestLegalMoves(t *testing.T) {
	type legalMoveTest struct {
		name         string
		pile, toPlay []deck.Card
		moves        []int
	}

	t.Run("four", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "four of ♣ beats two of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "four of ♠ does not beat five of ♣",
				pile: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Four, deck.Spades),
					deck.NewCard(deck.Six, deck.Spades),
					deck.NewCard(deck.Nine, deck.Hearts),
				},
				moves: []int{1, 2},
			},
			{
				name:   "four of ♥ does not beat King of ♣",
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{},
			},
			{
				name: "four of ♦ beats Seven of ♥",
				pile: []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Four, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
					deck.NewCard(deck.Ace, deck.Diamonds),
				},
				moves: []int{0, 1},
			},
			{
				name:   "four of ♣ does not beat Ace of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{},
			},
			{
				name:   "four of ♣ beats four of ♠",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "four of ♥ beats four of ♦",
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "four of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("ace", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "ace of ♠ does not beat seven of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Diamonds)},
				moves:  []int{},
			},
			{
				name:   "ace of ♠ beats king of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ace of ♦ beats Four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Diamonds)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ace of ♦ beats Ace of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Diamonds)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("seven", func(t *testing.T) {
		t.Run("to play", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "Seven of ♠ beats Four of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats five of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats six of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats seven of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ beats two of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♠ does not beat eight of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
				{
					name:   "seven of ♠ does not beat ace of ♣",
					pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Clubs)},
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
		t.Run("to beat", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "four of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "five of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "six of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "seven of ♣ beats seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "Two of ♣ beats Seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "eight of ♣ does not beat seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
				{
					name:   "ace of ♣ does not beat seven of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Ace, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Seven, deck.Spades)},
					moves:  []int{},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
	})

	t.Run("three", func(t *testing.T) {
		t.Run("three beats anything", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "three of ♣ beats two of ♦",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
					moves:  []int{0},
				},
				{
					name:   "three of ♠ beats five of ♣",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Spades)},
					pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					moves:  []int{0},
				},
				{
					name:   "three of ♥ beats King of ♣",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Hearts)},
					pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
					moves:  []int{0},
				},
				{
					name: "three of ♦ beats Seven of ♥",
					toPlay: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.Five, deck.Clubs),
					},
					pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
					moves: []int{0, 1},
				},
				{
					name:   "three of ♣ beats Ace of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
					moves:  []int{0},
				},
				{
					name:   "Three of ♣ beats four of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Three, deck.Clubs)},
					pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
					moves:  []int{0},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})

		t.Run("three on the pile is ignored", func(t *testing.T) {
			tt := []legalMoveTest{
				{
					name:   "four of ♣ does not beat six of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Four, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Six, deck.Hearts),
					},
					moves: []int{},
				},
				{
					name:   "Five of ♣ beats Four of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Four, deck.Spades),
					},
					moves: []int{0},
				},
				{
					name:   "Six of ♣ beats Five of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Six, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Five, deck.Hearts),
					},
					moves: []int{0},
				},
				{
					name:   "Seven of ♣ beats Five of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Seven, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Five, deck.Spades),
					},
					moves: []int{0},
				},
				{
					name:   "Two of ♣ beats Ace of ♦",
					toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Ace, deck.Diamonds),
					},
					moves: []int{0},
				},
				{
					name:   "Eight of ♣ does not beat Nine of ♠",
					toPlay: []deck.Card{deck.NewCard(deck.Eight, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Spades),
						deck.NewCard(deck.Nine, deck.Spades),
					},
					moves: []int{},
				},
				{
					name:   "Jack of ♣ does not beat King of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Jack, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.King, deck.Hearts),
					},
					moves: []int{},
				},
				{
					name:   "Queen of ♣ beats Jack of ♥",
					toPlay: []deck.Card{deck.NewCard(deck.Queen, deck.Clubs)},
					pile: []deck.Card{
						deck.NewCard(deck.Three, deck.Diamonds),
						deck.NewCard(deck.Jack, deck.Hearts),
					},
					moves: []int{0},
				},
			}

			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
				})
			}
		})
	})

	t.Run("two beats anything; anything beats a two", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "two of ♣ beats two of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♠ beats five of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "two of ♥ beats King of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "two of ♦ beats Seven of ♥",
				toPlay: []deck.Card{
					deck.NewCard(deck.Two, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
				},
				pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				moves: []int{0, 1},
			},
			{
				name:   "Two of ♣ does beats Ace of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "two of ♥ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "two of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Two, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("ten beats anything", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name:   "ten of ♣ beats two of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Two, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♠ beats five of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Spades)},
				pile:   []deck.Card{deck.NewCard(deck.Five, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♥ beats King of ♣",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.King, deck.Clubs)},
				moves:  []int{0},
			},
			{
				name: "ten of ♦ beats Seven of ♥",
				toPlay: []deck.Card{
					deck.NewCard(deck.Ten, deck.Diamonds),
					deck.NewCard(deck.Five, deck.Clubs),
				},
				pile:  []deck.Card{deck.NewCard(deck.Seven, deck.Hearts)},
				moves: []int{0, 1},
			},
			{
				name:   "ten of ♣ does not beat Ace of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Ace, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♠",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Spades)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♥ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Hearts)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♦",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Diamonds)},
				moves:  []int{0},
			},
			{
				name:   "ten of ♣ beats four of ♥",
				toPlay: []deck.Card{deck.NewCard(deck.Ten, deck.Clubs)},
				pile:   []deck.Card{deck.NewCard(deck.Four, deck.Hearts)},
				moves:  []int{0},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})

	t.Run("mixture", func(t *testing.T) {
		tt := []legalMoveTest{
			{
				name: "Three, Queen and King all beat an Eight",
				pile: []deck.Card{deck.NewCard(deck.Eight, deck.Spades)},
				toPlay: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.King, deck.Diamonds),
				},
				moves: []int{0, 1, 2},
			},
			{
				name: "Three and Ten beat a King; Queen does not",
				pile: []deck.Card{
					deck.NewCard(deck.King, deck.Diamonds),
					deck.NewCard(deck.Eight, deck.Spades),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Ten, deck.Diamonds),
				},
				moves: []int{0, 2},
			},
			{
				name: "Ten beats a King hidden by a Three; Jack and Queen do not",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Clubs),
					deck.NewCard(deck.King, deck.Diamonds),
					deck.NewCard(deck.Eight, deck.Spades),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Ten, deck.Diamonds),
				},
				moves: []int{2},
			},
			{
				name: "Jack, Three and Queen all beat an empty pile",
				pile: []deck.Card{},
				toPlay: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Spades),
				},
				moves: []int{0, 1, 2},
			},
			{
				name: "Three and Three beat a Queen; Jack does not",
				pile: []deck.Card{
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Three, deck.Hearts),
				},
				moves: []int{1, 2},
			},
			{
				name: "Three beats a Queen (under a Three); Seven and Jack do not",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Seven, deck.Diamonds),
					deck.NewCard(deck.Three, deck.Hearts),
				},
				moves: []int{2},
			},
			{
				name: "Jack, Jack and Seven cannot beat a Queen (hidden by 2 Threes)",
				pile: []deck.Card{
					deck.NewCard(deck.Three, deck.Hearts),
					deck.NewCard(deck.Three, deck.Spades),
					deck.NewCard(deck.Queen, deck.Diamonds),
				},
				toPlay: []deck.Card{
					deck.NewCard(deck.Jack, deck.Diamonds),
					deck.NewCard(deck.Seven, deck.Diamonds),
					deck.NewCard(deck.Jack, deck.Spades),
				},
				moves: []int{},
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				utils.AssertDeepEqual(t, getLegalMoves(tc.pile, tc.toPlay), tc.moves)
			})
		}
	})
}

func TestGameIsBurn(t *testing.T) {
	tt := []struct {
		name string
		pile []deck.Card
		want bool
	}{
		{
			"play a 10",
			combineCards(someCards(5),
				deck.NewCard(deck.Ten, deck.Hearts)),
			true,
		},
		{
			"four of the same suit (Four)",
			[]deck.Card{
				deck.NewCard(deck.Four, deck.Clubs),
				deck.NewCard(deck.Four, deck.Diamonds),
				deck.NewCard(deck.Four, deck.Hearts),
				deck.NewCard(deck.Four, deck.Spades),
			},
			true,
		},
		{
			"four of the same suit (Seven)",
			[]deck.Card{
				deck.NewCard(deck.Seven, deck.Clubs),
				deck.NewCard(deck.Seven, deck.Diamonds),
				deck.NewCard(deck.Seven, deck.Hearts),
				deck.NewCard(deck.Seven, deck.Spades),
			},
			true,
		},
		{
			"four of the same suit (Three)",
			[]deck.Card{
				deck.NewCard(deck.Three, deck.Clubs),
				deck.NewCard(deck.Three, deck.Diamonds),
				deck.NewCard(deck.Three, deck.Hearts),
				deck.NewCard(deck.Three, deck.Spades),
			},
			true,
		},
		{
			"four of the same suit (Two)",
			[]deck.Card{
				deck.NewCard(deck.Two, deck.Clubs),
				deck.NewCard(deck.Two, deck.Diamonds),
				deck.NewCard(deck.Two, deck.Hearts),
				deck.NewCard(deck.Two, deck.Spades),
			},
			true,
		},
		{
			"playing multiple Tens == one burn",
			combineCards(someCards(5),
				deck.NewCard(deck.Ten, deck.Hearts),
				deck.NewCard(deck.Ten, deck.Diamonds),
			),
			true,
		},
		{
			"threes are ignored",
			[]deck.Card{
				deck.NewCard(deck.Four, deck.Clubs),
				deck.NewCard(deck.Four, deck.Diamonds),
				deck.NewCard(deck.Three, deck.Hearts),
				deck.NewCard(deck.Four, deck.Hearts),
				deck.NewCard(deck.Three, deck.Spades),
				deck.NewCard(deck.Four, deck.Spades),
			},
			true,
		},
		{
			"three of the same suit (Five)",
			[]deck.Card{
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Five, deck.Diamonds),
				deck.NewCard(deck.Five, deck.Hearts),
				deck.NewCard(deck.Five, deck.Spades),
			},
			false,
		},
		{
			"two of the same suit (Seven)",
			[]deck.Card{
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Seven, deck.Diamonds),
				deck.NewCard(deck.Seven, deck.Hearts),
				deck.NewCard(deck.Seven, deck.Spades),
			},
			false,
		},
		{
			"four of the same suit (Five) separated by another card",
			[]deck.Card{
				deck.NewCard(deck.Five, deck.Diamonds),
				deck.NewCard(deck.Five, deck.Spades),
				deck.NewCard(deck.Nine, deck.Clubs),
				deck.NewCard(deck.Five, deck.Hearts),
				deck.NewCard(deck.Five, deck.Spades),
			},
			false,
		},
	}

	for _, tc := range tt[6:] {
		t.Run(tc.name, func(t *testing.T) {
			utils.AssertEqual(t, isBurn(tc.pile), tc.want)
		})
	}

	// pile has four of the same suit
	// play multiple 10s in one move == one burn
}
