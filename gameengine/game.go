package gameengine

import (
	"github.com/minaorangina/shed/deck"
)

// Stage represents the main stages in the game
type Stage int

const (
	handOrganisation Stage = iota
	clearDeck
	clearHand
)

func (s Stage) String() string {
	if s == 0 {
		return "handOrganisation"
	} else if s == 1 {
		return "clearDeck"
	} else if s == 2 {
		return "clearHand"
	}
	return ""
}

// Game represents a game
type Game struct {
	name    string
	engine  *GameEngine
	players *[]Player
	stage   Stage
	deck    deck.Deck
}

func makePlayers(playerInfo []playerInfo) ([]Player, error) {
	players := make([]Player, 0, len(playerInfo))
	for _, info := range playerInfo {
		p, playerErr := NewPlayer(info.id, info.name)
		if playerErr != nil {
			return []Player{}, playerErr
		}
		players = append(players, p)
	}

	return players, nil
}

// NewGame instantiates a new game of Shed
func NewGame(engine *GameEngine, playerInfo []playerInfo) *Game {
	cards := deck.New()
	players, _ := makePlayers(playerInfo)
	return &Game{name: "Shed", engine: engine, players: &players, deck: cards}
}

func (g *Game) start() {
	g.deck.Shuffle()
	g.dealHand()
	err := g.informPlayersAwaitReply()
	if err != nil {
		// handle error
	}
}

func (g *Game) dealHand() {
	for i := range *g.players {
		dealtHand := g.deck.Deal(3)
		dealtSeen := g.deck.Deal(3)
		dealtUnseen := g.deck.Deal(3)

		(*g.players)[i].hand = append((*g.players)[i].hand, dealtHand...)
		(*g.players)[i].seen = append((*g.players)[i].seen, dealtSeen...)
		(*g.players)[i].unseen = append((*g.players)[i].unseen, dealtUnseen...)
	}
}

func (g *Game) informPlayersAwaitReply() error {
	// construct opponents
	// construct object per player
	messages := make([]messageToPlayer, 0, len(*g.players))
	for _, p := range *g.players {
		o := buildOpponents(p.id, *g.players)
		m := g.buildMessageToPlayer(p, o, "Rearrange your hand")
		messages = append(messages, m)
	}
	// send on to game engine

	reorganised, err := g.engine.messagePlayersAwaitReply(messages)
	if err != nil {
		return err
	}
	_ = reorganised
	// wait for all responses to come back
	return nil
}

// Stage returns the game's current stage
func (g *Game) Stage() string {
	return g.stage.String()
}

// to test (easier when state hydration exists)
func (g *Game) buildMessageToPlayer(player Player, opponents []opponent, message string) messageToPlayer {
	return messageToPlayer{
		PlayState: g.engine.playState,
		GameStage: g.stage,
		PlayerID:  player.id,
		Message:   message,
		HandCards: player.cards().hand,
		SeenCards: player.cards().seen,
		Opponents: opponents,
	}
}

func buildOpponents(playerID string, players []Player) []opponent {
	opponents := []opponent{}
	for _, p := range players {
		if p.id == playerID {
			continue
		}
		opponents = append(opponents, opponent{
			ID: p.id, SeenCards: p.cards().seen,
		})
	}
	return opponents
}
