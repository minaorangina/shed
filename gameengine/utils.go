package gameengine

import (
	"os"
	"testing"

	"github.com/minaorangina/shed/players"
)

func somePlayers() players.Players {
	player1 := players.NewPlayer(players.NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := players.NewPlayer(players.NewID(), "Sally", os.Stdin, os.Stdout)
	players := players.Players([]*players.Player{player1, player2})
	return players
}

func gameEngineWithPlayers() GameEngine {
	ge, _ := New("theid", somePlayers(), nil)
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
	ps := []*players.Player{}
	for _, n := range names {
		player := players.NewPlayer(players.NewID(), n, os.Stdin, os.Stdout)
		ps = append(ps, player)
	}

	return players.Players(ps)
}

func playersToNames(players players.Players) []string {
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name)
	}

	return names
}

func playerInfoToPlayers(playerInfo []playerInfo) players.Players {
	ps := []*players.Player{}
	for _, info := range playerInfo {
		ps = append(ps, players.NewPlayer(info.id, info.name, os.Stdin, os.Stdout))
	}

	return players.Players(ps)
}

func charsUnique(s string) bool {
	seen := map[string]bool{}
	for _, c := range s {
		if _, ok := seen[string(c)]; ok {
			return false
		}
		seen[string(c)] = true
	}
	return true
}

func charsInRange(chars string, lower, upper int) bool {
	for _, char := range chars {
		if int(char) < lower || int(char) > upper {
			return false
		}
	}
	return true
}

func assertStringEquality(t *testing.T, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("got %s, want %s", got, want)
	}
}
