package gameengine

import (
	"os"
	"testing"
)

func buildOpponents(playerID string, players Players) []opponent {
	opponents := []opponent{}
	for _, p := range players {
		if p.ID == playerID {
			continue
		}
		opponents = append(opponents, opponent{
			ID: p.ID, Seen: p.cards().seen, Name: p.Name,
		})
	}
	return opponents
}

func messagesToInitialCards(messages []InboundMessage) map[string]initialCards {
	reorganised := map[string]initialCards{}

	for _, msg := range messages {
		reorganised[msg.PlayerID] = initialCards{
			seen: msg.Seen,
			hand: msg.Hand,
		}
	}

	return reorganised
}

func namesToPlayers(names []string) Players {
	players := []*Player{}
	for _, n := range names {
		player := NewPlayer(NewID(), n, os.Stdin, os.Stdout)
		players = append(players, player)
	}

	return Players(players)
}

func playersToNames(players Players) []string {
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name)
	}

	return names
}

func playerInfoToPlayers(playerInfo []playerInfo) Players {
	players := []*Player{}
	for _, info := range playerInfo {
		players = append(players, NewPlayer(info.id, info.name, os.Stdin, os.Stdout))
	}

	return Players(players)
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
