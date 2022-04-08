package engine

import (
	"os"
	"sync"

	"github.com/minaorangina/shed/game"
	"github.com/minaorangina/shed/protocol"
)

type SpyGame struct {
	startCalled bool
	mu          *sync.Mutex
}

func NewSpyGame() *SpyGame {
	return &SpyGame{mu: &sync.Mutex{}}
}

func (g *SpyGame) AwaitingResponse() protocol.Cmd {
	return protocol.Null
}

func (g *SpyGame) Start(info []protocol.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.startCalled = true
	return nil
}

func (g *SpyGame) Next() ([]protocol.OutboundMessage, error) {
	return nil, nil
}

func (g *SpyGame) ReceiveResponse(messages []protocol.InboundMessage) ([]protocol.OutboundMessage, error) {
	return nil, nil
}

func (g *SpyGame) GameOver() bool {
	return false
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

func playerInfoToPlayers(playerInfo []protocol.Player) Players {
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
		Game:      game.ExistingShed(game.ShedOpts{}),
	})
	return ge
}

func buildOpponents(playerID string, playerCards map[string]*game.PlayerCards) []protocol.Opponent {
	opponents := []protocol.Opponent{}
	for id, pc := range playerCards {
		if id != playerID {
			opponents = append(opponents, protocol.Opponent{
				PlayerID: id, Seen: pc.Seen,
			})
		}
	}
	return opponents
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
