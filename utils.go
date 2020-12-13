package shed

import (
	"os"
)

type SpyGame struct {
	startCalled bool
}

func (g *SpyGame) Start(playerIDs []string) error {
	g.startCalled = true
	return nil
}

func (g *SpyGame) Next() ([]OutboundMessage, bool) {
	return nil, false
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
	ge := NewGameEngine(GameEngineOpts{
		GameID:    "theid",
		CreatorID: "some-user-id",
		Players:   SomePlayers()})
	return ge
}

func buildOpponents(playerID string, ps Players) []Opponent {
	opponents := []Opponent{}
	for _, p := range ps {
		if p.ID() == playerID {
			continue
		}
		opponents = append(opponents, Opponent{
			ID: p.ID(), Seen: p.Cards().Seen, Name: p.Name(),
		})
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
