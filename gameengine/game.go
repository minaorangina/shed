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

func namesToPlayers(names []string) ([]Player, error) {
	players := make([]Player, 0, len(names))
	for i, name := range names {
		p, playerErr := NewPlayer(i, name)
		if playerErr != nil {
			return []Player{}, playerErr
		}
		players = append(players, p)
	}

	return players, nil
}

// NewGame instantiates a new game of Shed
func NewGame(engine *GameEngine, playerNames []string) *Game {
	cards := deck.New()
	players, _ := namesToPlayers(playerNames)
	return &Game{name: "Shed", engine: engine, players: &players, deck: cards}
}

func (g *Game) start() {
	g.deck.Shuffle()
	g.dealHand()
	g.informPlayersAwaitReply()
}

func (g *Game) dealHand() {
	for _, p := range *g.players {
		dealtHand := g.deck.Deal(3)
		dealtSeen := g.deck.Deal(3)
		dealtUnseen := g.deck.Deal(3)

		p.cards.hand = &dealtHand
		p.cards.seen = &dealtSeen
		p.cards.unseen = &dealtUnseen
	}
}

func (g *Game) informPlayersAwaitReply() {
	// construct opponents
	// construct object per player
	messages := make([]messageToPlayer, 0, len(*g.players))
	for _, p := range *g.players {
		o := buildOpponents(p.id, *g.players)
		m := g.buildMessageToPlayer(p, o, "Rearrange your hand")
		messages = append(messages, m)
	}
	// send on to game engine
	// wait for all responses to come back
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
		HandCards: *player.cards.hand,
		SeenCards: *player.cards.seen,
		Opponents: opponents,
	}
}

func buildOpponents(playerID int, players []Player) []opponent {
	opponents := []opponent{}
	for id, p := range players {
		if id == playerID {
			continue
		}
		var seen []deck.Card
		if p.cards.seen == nil {
			seen = nil
		} else {
			seen = *p.cards.seen
		}
		opponents = append(opponents, opponent{
			ID: id, Name: p.name, SeenCards: seen,
		})
	}
	return opponents
}
