package shed

import (
	"errors"

	"github.com/minaorangina/shed/deck"
)

// Stage represents the main stages in the game
type Stage int

const (
	clearDeck Stage = iota
	clearCards
)

const (
	minPlayers = 2
	maxPlayers = 4
)

func (s Stage) String() string { // TODO: test
	if s == 0 {
		return "clearDeck"
	} else if s == 1 {
		return "clearCards"
	}
	return ""
}

type Game interface {
	Start(playerIDs []string) error
	Next() (messages []OutboundMessage, err error)
}

type shed struct {
	deck        deck.Deck
	pile        []deck.Card
	playerCards map[string]*PlayerCards
	playerIDs   []string
	currentTurn int
	stage       Stage
}

type ShedOpts struct {
	deck        deck.Deck
	pile        []deck.Card
	playerCards map[string]*PlayerCards
	playerIDs   []string
	currentTurn int
	stage       Stage
}

// NewShed constructs a new game of Shed
func NewShed(opts ShedOpts) *shed {
	s := &shed{
		deck:        opts.deck,
		pile:        opts.pile,
		playerCards: opts.playerCards,
		playerIDs:   opts.playerIDs,
		currentTurn: opts.currentTurn,
		stage:       opts.stage,
	}
	if s.deck == nil {
		s.deck = deck.New()
	}
	if s.pile == nil {
		s.pile = []deck.Card{}
	}
	if s.playerCards == nil {
		s.playerCards = map[string]PlayerCards{}
	}
	if s.playerIDs == nil {
		s.playerIDs = []string{}
	}

	return s
}

func (s *shed) Start(playerIDs []string) error {
	if s == nil {
		return ErrNilGame
	}
	if len(playerIDs) < minPlayers {
		return ErrTooFewPlayers
	}
	if len(playerIDs) > maxPlayers {
		return ErrTooManyPlayers
	}

	cards := map[string]PlayerCards{}
	for _, id := range playerIDs {
		cards[id] = PlayerCards{}
	}
	s.gamePlayers = cards

	return nil
}

func (s *shed) Next() ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.gamePlayers == nil {
		return nil, errors.New("game has no players")
	}
	return nil, nil
}
