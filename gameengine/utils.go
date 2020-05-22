package gameengine

import "os"

func buildOpponents(playerID string, players AllPlayers) []opponent {
	opponents := []opponent{}
	for id, p := range players {
		if id == playerID {
			continue
		}
		opponents = append(opponents, opponent{
			ID: p.ID, Seen: p.cards().seen, Name: p.Name,
		})
	}
	return opponents
}

func messagesToInitialCards(messages InboundMessages) map[string]initialCards {
	reorganised := map[string]initialCards{}

	for id, msg := range messages {
		reorganised[id] = initialCards{
			seen: msg.Seen,
			hand: msg.Hand,
		}
	}

	return reorganised
}

func namesToAllPlayers(names []string) AllPlayers {
	players := []Player{}
	for _, n := range names {
		players = append(players, NewPlayer(NewID(), n, os.Stdin, os.Stdout))
	}

	return NewAllPlayers(players...)
}

func allPlayersToNames(players AllPlayers) []string {
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name)
	}

	return names
}

func playerInfoToAllPlayers(playerInfo []playerInfo) AllPlayers {
	players := []Player{}
	for _, info := range playerInfo {
		players = append(players, NewPlayer(info.id, info.name, os.Stdin, os.Stdout))
	}

	return NewAllPlayers(players...)
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
