package shed

import "github.com/minaorangina/shed/deck"

func getLegalMoves(pile, toPlay []deck.Card) []int {
	candidates := []deck.Card{}
	for _, c := range pile {
		if c.Rank != deck.Three {
			candidates = append(candidates, c)
		}
	}

	moves := map[int]struct{}{}

	// can play any card on an empty pile
	if len(candidates) == 0 {
		for i := range toPlay {
			moves[i] = struct{}{}
		}

		return setToIntSlice(moves)
	}

	// tens (and twos and threes) beat anything
	for i, tp := range toPlay {
		if tp.Rank == deck.Ten {
			moves[i] = struct{}{}
			continue
		}
	}

	topmostCard := candidates[0]
	// two
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
		// skip tens (and twos and threes)
		if tp.Rank == deck.Ten {
			continue
		}

		tpValue := cardValues[tp.Rank]
		pileValue := cardValues[topmostCard.Rank]

		if tpValue >= pileValue {
			moves[i] = struct{}{}
		}
	}

	return setToIntSlice(moves)
}
