package shed

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/minaorangina/shed/players"
)

func APlayer(id, name string) *players.Player {
	return players.NewPlayer(id, name, &bytes.Buffer{}, ioutil.Discard)
}

func SomePlayers() players.Players {
	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)
	players := players.NewPlayers(player1, player2)
	return players
}

func gameEngineWithPlayers() GameEngine {
	ge, _ := New("theid", SomePlayers(), nil)
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
