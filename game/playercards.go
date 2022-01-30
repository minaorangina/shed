package game

import "github.com/minaorangina/shed/deck"

type PlayerCards struct {
	Hand, Seen, Unseen []deck.Card
	UnseenVisibility   map[deck.Card]bool
}

func NewPlayerCards(
	hand, seen, unseen []deck.Card,
	unseenVisibility map[deck.Card]bool,
) *PlayerCards {
	if hand == nil {
		hand = []deck.Card{}
	}
	if seen == nil {
		seen = []deck.Card{}
	}
	if unseen == nil {
		unseen = []deck.Card{}
	}
	if unseenVisibility == nil {
		unseenVisibility = map[deck.Card]bool{}
		for _, c := range unseen {
			unseenVisibility[c] = false
		}
	}

	pc := &PlayerCards{
		Hand:             hand,
		Seen:             seen,
		Unseen:           unseen,
		UnseenVisibility: unseenVisibility,
	}

	// enforce validity

	return pc
}

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
