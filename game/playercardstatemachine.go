package game

import "github.com/minaorangina/shed/deck"

func playerCardsValid(cards *PlayerCards) bool {
	if cards == nil {
		return false
	}

	if len(cards.Seen) > numCardsInGroup {
		return false
	}

	if len(cards.Unseen) > numCardsInGroup {
		return false
	}

	if len(cards.Unseen) < numCardsInGroup && len(cards.Seen) != 0 {
		return false
	}

	visited := map[deck.Card]struct{}{}

	for _, c := range cards.Hand {
		if _, ok := visited[c]; ok {
			return false
		}
		visited[c] = struct{}{}
	}
	for _, c := range cards.Seen {
		if _, ok := visited[c]; ok {
			return false
		}
		visited[c] = struct{}{}
	}
	for _, c := range cards.Unseen {
		if _, ok := visited[c]; ok {
			return false
		}
		visited[c] = struct{}{}
	}

	return true
}
