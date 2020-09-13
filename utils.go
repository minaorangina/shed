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
