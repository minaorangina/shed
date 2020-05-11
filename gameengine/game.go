package gameengine

import (
	"fmt"

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
	players map[string]*Player
	stage   Stage
	deck    deck.Deck
}

func makePlayers(playerInfo []playerInfo) (map[string]*Player, error) {
	players := make(map[string]*Player)
	for _, info := range playerInfo {
		p := NewPlayer(info.id, info.name)
		players[p.id] = &p
	}

	return players, nil
}

// NewGame instantiates a new game of Shed
func NewGame(engine *GameEngine, playerInfo []playerInfo) (*Game, error) {
	if len(playerInfo) < 2 {
		return nil, fmt.Errorf("Could not construct Game: minimum of 2 players required (supplied %d)", len(playerInfo))
	}
	if len(playerInfo) > 4 {
		return nil, fmt.Errorf("Could not construct Game: maximum of 4 players allowed (supplied %d)", len(playerInfo))
	}
	cards := deck.New()
	players, _ := makePlayers(playerInfo)
	return &Game{name: "Shed", engine: engine, players: players, deck: cards}, nil
}

func (g *Game) start() {
	g.deck.Shuffle()
	g.dealHand()
	err := g.messagePlayersAwaitReply()
	if err != nil {
		// handle error
	}
}

func (g *Game) dealHand() {
	for _, p := range g.players {
		dealtHand := g.deck.Deal(3)
		dealtSeen := g.deck.Deal(3)
		dealtUnseen := g.deck.Deal(3)

		p.hand = dealtHand
		p.seen = dealtSeen
		p.unseen = dealtUnseen
	}
}

func (g *Game) messagePlayersAwaitReply() error {
	messages := make(map[string]messageToPlayer)
	for _, p := range g.players {
		o := buildOpponents(p.id, g.players)
		m := g.buildMessageToPlayer(p, o, "Rearrange your hand")
		messages[p.id] = m
	}

	// send on to game engine
	reply, err := g.engine.messagePlayersAwaitReply(messages)
	if err != nil {
		return err
	}
	reorganised := mapMessagesToHand(reply)
	for id, p := range g.players {
		p.hand = reorganised[id].HandCards
		p.seen = reorganised[id].SeenCards
	}

	return nil
}

// Stage returns the game's current stage
func (g *Game) Stage() string {
	return g.stage.String()
}

// to test (easier when state hydration exists)
func (g *Game) buildMessageToPlayer(player *Player, opponents []opponent, message string) messageToPlayer {
	return messageToPlayer{
		PlayState: g.engine.playState,
		GameStage: g.stage,
		PlayerID:  player.id,
		Name:      player.name,
		Message:   message,
		HandCards: player.cards().hand,
		SeenCards: player.cards().seen,
		Opponents: opponents,
	}
}

func buildOpponents(playerID string, players map[string]*Player) []opponent {
	opponents := []opponent{}
	for id, p := range players {
		if id == playerID {
			continue
		}
		opponents = append(opponents, opponent{
			ID: p.id, SeenCards: p.cards().seen, Name: p.name,
		})
	}
	return opponents
}

func mapMessagesToHand(messages map[string]messageFromPlayer) map[string]reorganisedHand {
	reorganised := map[string]reorganisedHand{}

	for id, msg := range messages {
		reorganised[id] = reorganisedHand{
			SeenCards: msg.SeenCards,
			HandCards: msg.HandCards,
		}
	}

	return reorganised
}
