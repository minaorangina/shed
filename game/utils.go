package game

import (
	"sort"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
)

func intSliceToSet(s []int) map[int]struct{} {
	set := map[int]struct{}{}
	for i := range s {
		set[i] = struct{}{}
	}

	return set
}

func setToIntSlice(set map[int]struct{}) []int {
	s := []int{}
	for key := range set {
		s = append(s, key)
	}

	sort.Ints(s)

	return s
}

func cardSliceToSet(s []deck.Card) map[deck.Card]struct{} {
	set := map[deck.Card]struct{}{}
	for _, key := range s {
		set[key] = struct{}{}
	}
	return set
}

func setToCardSlice(set map[deck.Card]struct{}) []deck.Card {
	s := []deck.Card{}
	for key := range set {
		s = append(s, key)
	}
	return s
}

func cardsUnique(cards []deck.Card) bool {
	seen := map[deck.Card]struct{}{}
	for _, c := range cards {
		if _, ok := seen[c]; ok {
			return false
		}
		seen[c] = struct{}{}
	}
	return true
}

func someDeck(num int) deck.Deck {
	d := deck.New()
	d.Shuffle()
	return deck.Deck(d.Deal(num))
}

func someCards(num int) []deck.Card {
	d := someDeck(num)
	return []deck.Card(d)
}

func combineCards(cards []deck.Card, toAdd ...deck.Card) []deck.Card {
	for _, c := range toAdd {
		cards = append(cards, c)
	}

	return cards
}

func somePlayerCards(num int) *PlayerCards {
	unseen := someDeck(num)
	pc := NewPlayerCards(
		someDeck(num),
		someDeck(num),
		unseen,
		nil,
	)
	for _, c := range unseen {
		pc.UnseenVisibility[c] = false
	}

	return pc
}

func containsCard(s []deck.Card, targets ...deck.Card) bool {
	for _, c := range s {
		for _, tg := range targets {
			if c == tg {
				return true
			}
		}
	}
	return false
}

func sliceContainsPlayerID(haystack []protocol.Player, needle string) bool {
	var found bool
	for _, h := range haystack {
		if needle == h.PlayerID {
			found = true
			break
		}
	}

	return found
}
