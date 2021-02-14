package shed

import (
	"github.com/minaorangina/shed/deck"
)

const (
	minPlayers = 2
	maxPlayers = 4
	burnNum    = 4
)

var cardValues = map[deck.Rank]int{
	deck.Four:  0,
	deck.Five:  1,
	deck.Six:   2,
	deck.Eight: 4,
	deck.Nine:  5,
	deck.Jack:  6,
	deck.Queen: 7,
	deck.King:  8,
	deck.Ace:   9,
	// special powers
	deck.Two:   9,
	deck.Three: 9,
	deck.Seven: 3,
	deck.Ten:   9,
}

var sevenBeaters = map[deck.Rank]bool{
	deck.Four:  true,
	deck.Five:  true,
	deck.Six:   true,
	deck.Eight: false,
	deck.Nine:  false,
	deck.Jack:  false,
	deck.Queen: false,
	deck.King:  false,
	deck.Ace:   false,
	// special powers
	deck.Two:   true,
	deck.Three: true,
	deck.Seven: true,
}

func getLegalMoves(pile, toPlay []deck.Card) []int {
	pileWithoutThrees := []deck.Card{}
	// Filter out Threes
	for _, c := range pile {
		if c.Rank != deck.Three {
			pileWithoutThrees = append(pileWithoutThrees, c)
		}
	}

	moves := map[int]struct{}{}

	// Can play any card on an empty pile
	if len(pileWithoutThrees) == 0 {
		for i := range toPlay {
			moves[i] = struct{}{}
		}

		return setToIntSlice(moves)
	}

	// Tens beat anything
	for i, tp := range toPlay {
		if tp.Rank == deck.Ten {
			moves[i] = struct{}{}
			continue
		}
	}

	topmostCard := pileWithoutThrees[len(pileWithoutThrees)-1]

	// Can play any card on a Two
	if topmostCard.Rank == deck.Two {
		for i := range toPlay {
			moves[i] = struct{}{}
		}
		return setToIntSlice(moves)
	}

	// seven
	if topmostCard.Rank == deck.Seven {
		for i, tp := range toPlay {
			if wins := sevenBeaters[tp.Rank]; wins {
				moves[i] = struct{}{}
			}
		}
		return setToIntSlice(moves)
	}

	for i, tp := range toPlay {
		// skip tens (already checked)
		if tp.Rank == deck.Ten {
			continue
		}

		tpValue := cardValues[tp.Rank]
		topmostCardValue := cardValues[topmostCard.Rank]
		if tpValue >= topmostCardValue {
			moves[i] = struct{}{}
		}
	}

	return setToIntSlice(moves)
}

func isBurn(pile []deck.Card) bool {
	if len(pile) < burnNum {
		return false
	}

	// Here, the most recently played card is at index len-1
	topCard := pile[len(pile)-1]
	if topCard.Rank == deck.Ten {
		return true
	}

	// Check for Three burn
	numThrees := 0
	for _, c := range pile[len(pile)-burnNum:] {
		if c.Rank == deck.Three {
			numThrees++
		}
	}
	if numThrees == burnNum {
		return true
	}

	// Take the topmost cards, excluding Threes
	// Will be 4 cards or fewer
	topCards := []deck.Card{}
	for i := len(pile) - 1; i >= 0; i-- {
		if len(topCards) == burnNum {
			break
		}
		card := pile[i]
		if card.Rank != deck.Three {
			topCards = append(topCards, card)
		}
	}

	// Reject fewer than 4 cards
	if len(topCards) < burnNum {
		return false
	}

	// From this point the cards are reversed
	// with the most recently played card at index 0
	topCard = topCards[0]
	for _, c := range topCards {
		if c.Rank != topCard.Rank {
			return false
		}
	}

	return true
}
