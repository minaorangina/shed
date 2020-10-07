package shed

import (
	"github.com/minaorangina/shed/players"
)

func gameEngineWithPlayers() GameEngine {
	ge, _ := New("theid", players.SomePlayers(), nil)
	return ge
}

func buildOpponents(playerID string, ps players.Players) []players.Opponent {
	opponents := []players.Opponent{}
	for _, p := range ps {
		if p.ID == playerID {
			continue
		}
		opponents = append(opponents, players.Opponent{
			ID: p.ID, Seen: p.Cards().Seen, Name: p.Name,
		})
	}
	return opponents
}
