package shed

import (
	"os"
	"sort"

	"github.com/minaorangina/shed/deck"
)

type SpyGame struct {
	startCalled bool
}

func (g *SpyGame) Start(playerIDs []string) error {
	g.startCalled = true
	return nil
}

func (g *SpyGame) Next() ([]OutboundMessage, error) {
	return nil, nil
}

func (g *SpyGame) ReceiveResponse(messages []InboundMessage) ([]OutboundMessage, error) {
	return nil, nil
}

func messagesToInitialCards(messages []InboundMessage) map[string]InitialCards {
	reorganised := map[string]InitialCards{}

	for _, msg := range messages {
		reorganised[msg.PlayerID] = InitialCards{
			Seen: msg.Seen,
			Hand: msg.Hand,
		}
	}

	return reorganised
}

func namesToPlayers(names []string) Players {
	ps := []Player{}
	for _, n := range names {
		player := NewTestPlayer(NewID(), n, os.Stdin, os.Stdout)
		ps = append(ps, player)
	}

	return ps
}

func playersToNames(players Players) []string {
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name())
	}

	return names
}

func playerInfoToPlayers(playerInfo []PlayerInfo) Players {
	ps := []Player{}
	for _, info := range playerInfo {
		ps = append(ps, NewTestPlayer(info.PlayerID, info.Name, os.Stdin, os.Stdout))
	}

	return NewPlayers(ps...)
}

func gameEngineWithPlayers() GameEngine {
	ge, _ := NewGameEngine(GameEngineOpts{
		GameID:    "theid",
		CreatorID: "some-user-id",
		Players:   SomePlayers(),
		Game:      NewShed(ShedOpts{}),
	})
	return ge
}

func buildOpponents(playerID string, playerCards map[string]*PlayerCards) []Opponent {
	opponents := []Opponent{}
	for id, pc := range playerCards {
		if id != playerID {
			opponents = append(opponents, Opponent{
				ID: id, Seen: pc.Seen,
			})
		}
	}
	return opponents
}

// NewTestGameStore is a convenience function for creating InMemoryGameStore in tests
func NewTestGameStore(
	games map[string]GameEngine,
	pendingPlayers map[string][]PlayerInfo,
) *InMemoryGameStore {
	if games == nil {
		games = map[string]GameEngine{}
	}

	if pendingPlayers == nil {
		pendingPlayers = map[string][]PlayerInfo{}
	}

	return &InMemoryGameStore{
		Games:          games,
		PendingPlayers: pendingPlayers,
	}
}

func setToIntSlice(set map[int]struct{}) []int {
	s := []int{}
	for key := range set {
		s = append(s, key)
	}

	sort.Ints(s)

	return s
}

func setToCardSlice(set map[deck.Card]struct{}) []deck.Card {
	s := []deck.Card{}
	for key := range set {
		s = append(s, key)
	}
	return s
}

func cardSliceToSet(s []deck.Card) map[deck.Card]struct{} {
	set := map[deck.Card]struct{}{}
	for _, key := range s {
		set[key] = struct{}{}
	}
	return set
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

func sliceContainsString(haystack []string, needle string) bool {
	var found bool
	for _, h := range haystack {
		if needle == h {
			found = true
			break
		}
	}

	return found
}
