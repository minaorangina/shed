package game

import (
	"testing"

	"github.com/minaorangina/shed/deck"
	"github.com/stretchr/testify/assert"
)

// ideally these would be table tests.
// good candidate for declarative way of defining player cards
func TestPlayerCardsValid(t *testing.T) {
	t.Run("seen cards must not exceed 3", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(3),
			d.Deal(4), // randomise?
			d.Deal(3),
			nil,
		)
		assert.False(t, playerCardsValid(pc))
	})

	t.Run("unseen cards must not exceed 3", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(3),
			d.Deal(3),
			d.Deal(4), // randomise?
			nil,
		)
		assert.False(t, playerCardsValid(pc))
	})

	t.Run("if unseen cards < 3, seen cards must be zero", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(0),
			d.Deal(1),
			d.Deal(2),
			nil,
		)
		assert.False(t, playerCardsValid(pc))
	})

	t.Run("cards must be unique", func(t *testing.T) {
		t.Run("duplicates in hand", func(t *testing.T) {
			d := deck.New()
			pc := NewPlayerCards(
				[]deck.Card{
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Nine, deck.Hearts),
				},
				d.Deal(3),
				d.Deal(3),
				nil,
			)
			assert.False(t, playerCardsValid(pc))
		})

		t.Run("duplicates in seen", func(t *testing.T) {
			d := deck.New()
			pc := NewPlayerCards(
				d.Deal(3),
				[]deck.Card{
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Nine, deck.Hearts),
				},
				d.Deal(3),
				nil,
			)
			assert.False(t, playerCardsValid(pc))
		})

		t.Run("duplicates in unseen", func(t *testing.T) {
			d := deck.New()
			pc := NewPlayerCards(
				d.Deal(3),
				d.Deal(3),
				[]deck.Card{
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Ace, deck.Spades),
					deck.NewCard(deck.Nine, deck.Hearts),
				},
				nil,
			)
			assert.False(t, playerCardsValid(pc))
		})

		t.Run("duplicates across card sets", func(t *testing.T) {
			d := deck.New()
			h := append(d.Deal(2), deck.NewCard(deck.Ace, deck.Spades))
			u := append(d.Deal(2), deck.NewCard(deck.Ace, deck.Spades))

			pc := NewPlayerCards(
				h,
				d.Deal(3),
				u,
				nil,
			)
			assert.False(t, playerCardsValid(pc))
		})
	})

	t.Run("accepts fresh set", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(3),
			d.Deal(3),
			d.Deal(3),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})

	t.Run("accepts hand > 3", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(17),
			d.Deal(3),
			d.Deal(3),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})

	t.Run("accepts only unseen cards", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(0),
			d.Deal(0),
			d.Deal(3),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})

	t.Run("accepts only hand cards", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(2),
			d.Deal(0),
			d.Deal(0),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})

	t.Run("accepts only hand and unseen cards", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(2),
			d.Deal(0),
			d.Deal(3),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})

	t.Run("accepts only seen and unseen cards", func(t *testing.T) {
		d := deck.New()
		pc := NewPlayerCards(
			d.Deal(0),
			d.Deal(2),
			d.Deal(3),
			nil,
		)
		assert.True(t, playerCardsValid(pc))
	})
}
