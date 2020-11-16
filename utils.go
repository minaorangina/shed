package shed

import (
	"os"

	"github.com/minaorangina/shed/players"
)

func messagesToInitialCards(messages []players.InboundMessage) map[string]players.InitialCards {
	reorganised := map[string]players.InitialCards{}

	for _, msg := range messages {
		reorganised[msg.PlayerID] = players.InitialCards{
			Seen: msg.Seen,
			Hand: msg.Hand,
		}
	}

	return reorganised
}

func namesToPlayers(names []string) players.Players {
	ps := []players.Player{}
	for _, n := range names {
		player := players.NewTestPlayer(players.NewID(), n, os.Stdin, os.Stdout)
		ps = append(ps, player)
	}

	return ps
}

func playersToNames(players players.Players) []string {
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name())
	}

	return names
}

func playerInfoToPlayers(playerInfo []PlayerInfo) players.Players {
	ps := []players.Player{}
	for _, info := range playerInfo {
		ps = append(ps, players.NewTestPlayer(info.PlayerID, info.Name, os.Stdin, os.Stdout))
	}

	return players.NewPlayers(ps...)
}

func gameEngineWithPlayers() GameEngine {
	ge, _ := New("theid", "some-user-id", players.SomePlayers(), nil)
	return ge
}

func buildOpponents(playerID string, ps players.Players) []players.Opponent {
	opponents := []players.Opponent{}
	for _, p := range ps {
		if p.ID() == playerID {
			continue
		}
		opponents = append(opponents, players.Opponent{
			ID: p.ID(), Seen: p.Cards().Seen, Name: p.Name(),
		})
	}
	return opponents
}

// NewTestGameStore is a convenience function for creating InMemoryGameStore in tests
func NewTestGameStore(
	activeGames,
	inactiveGames map[string]GameEngine,
	pendingPlayers map[string][]PlayerInfo,
) *InMemoryGameStore {
	if activeGames == nil {
		activeGames = map[string]GameEngine{}
	}
	if inactiveGames == nil {
		inactiveGames = map[string]GameEngine{}
	}
	if pendingPlayers == nil {
		pendingPlayers = map[string][]PlayerInfo{}
	}

	return &InMemoryGameStore{
		ActiveGames:    activeGames,
		InactiveGames:  inactiveGames,
		PendingPlayers: pendingPlayers,
	}
}
