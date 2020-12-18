package shed

import (
	"errors"
	"fmt"

	"github.com/minaorangina/shed/protocol"

	"github.com/minaorangina/shed/deck"
)

// Stage represents the main stages in the game
type Stage int

const (
	preGame Stage = iota
	clearDeck
	clearCards
)

const (
	minPlayers = 2
	maxPlayers = 4
)

var cardValues = map[deck.Rank]int{
	deck.Four:  0,
	deck.Five:  1,
	deck.Six:   2,
	deck.Eight: 4,
	deck.Nine:  5,
	deck.Jack:  6,
	deck.Queen: 7,
	deck.King:  8,
	deck.Ace:   9,
	// special powers
	deck.Two:   9,
	deck.Three: 9,
	deck.Seven: 3,
	deck.Ten:   9,
}

var sevenBeaters = map[deck.Rank]bool{
	deck.Four:  true,
	deck.Five:  true,
	deck.Six:   true,
	deck.Eight: false,
	deck.Nine:  false,
	deck.Jack:  false,
	deck.Queen: false,
	deck.King:  false,
	deck.Ace:   false,
	// special powers
	deck.Two:   true,
	deck.Three: true,
	deck.Seven: true,
}

type Game interface {
	Start(playerIDs []string) error
	Next() ([]OutboundMessage, error)
	ReceiveResponse([]InboundMessage) ([]OutboundMessage, error)
}

type shed struct {
	deck             deck.Deck
	pile             []deck.Card
	playerCards      map[string]*PlayerCards
	playerIDs        []string
	currentTurn      int
	stage            Stage
	awaitingResponse bool
}

type ShedOpts struct {
	deck             deck.Deck
	pile             []deck.Card
	playerCards      map[string]*PlayerCards
	playerIDs        []string
	currentTurn      int
	stage            Stage
	awaitingResponse bool
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
		s.playerCards = map[string]*PlayerCards{}
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

	// deal cards etc

	cards := map[string]*PlayerCards{}
	for _, id := range playerIDs {
		cards[id] = &PlayerCards{
			Hand: []deck.Card{},
			Seen: []deck.Card{},
		}
	}

	s.playerIDs = playerIDs
	s.playerCards = cards
	s.stage = clearDeck

	return nil
}

func (s *shed) Next() ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.playerCards == nil {
		return nil, ErrNoPlayers
	}

	switch s.stage {
	case clearDeck:
		playerID := s.playerIDs[s.currentTurn]
		playerCards := s.playerCards[playerID]

		// can player play? if so, emit event and await response
		if true {
			s.awaitingResponse = true

			return []OutboundMessage{{
				PlayerID:       playerID,
				Command:        protocol.PlayHand,
				Hand:           s.playerCards[playerID].Hand,
				Seen:           s.playerCards[playerID].Seen,
				ExpectResponse: true,
			}}, nil
		}
		// if not, collect pile and move on

		playerCards.Hand = append(s.playerCards[playerID].Hand, s.pile...)

		s.pile = []deck.Card{}

		return []OutboundMessage{{
			PlayerID: playerID,
			Command:  protocol.NoLegalMoves,
			Hand:     s.playerCards[playerID].Hand,
			Seen:     s.playerCards[playerID].Seen,
		}}, nil
	}

	return nil, nil
}

func (s *shed) ReceiveResponse(msgs []InboundMessage) ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.playerCards == nil {
		return nil, ErrNoPlayers
	}

	if s.stage == preGame {

	}

	if s.stage == clearDeck {
		if len(msgs) != 1 {
			return nil, fmt.Errorf("expected one message, got %d", len(msgs))
		}

		msg := msgs[0]
		playerID := msg.PlayerID // check it's an id we recognise

		switch msg.Command {
		case protocol.PlayHand:
			// check this is a legal move. this has already been done, but worth
			// double checking in case of client tampering.

			// add cards to pile
			toPile := []deck.Card{}
			for _, cardIdx := range msg.Decision {
				// copy card from Hand
				toPile = append(toPile, s.playerCards[playerID].Hand[cardIdx])
			}

			s.pile = append(s.pile, toPile...)

			// remove selected cards from hand
			newHand := []deck.Card{}
			for _, hc := range s.playerCards[playerID].Hand {
				for _, pc := range toPile {
					if hc != pc {
						newHand = append(newHand, hc)
					}
				}
			}

			s.playerCards[playerID].Hand = newHand

			// pluck from deck
			fromDeck := s.deck.Deal(len(msg.Decision))
			s.playerCards[playerID].Hand = append(s.playerCards[playerID].Hand, fromDeck...)

			// return messages with no response expected.
			toSend := []OutboundMessage{{
				PlayerID: playerID,
				Command:  protocol.Replenish,
				Hand:     newHand,
				Pile:     s.pile,
			}}

			for _, id := range s.playerIDs {
				if id != playerID {
					toSend = append(toSend, s.buildEndOfTurnMessage(id))
				}
			}

			s.awaitingResponse = false
			return toSend, nil
		}

	}
	// stage 2
	if s.stage == clearCards {

	}

	return nil, errors.New("invalid game state")
}

func (s *shed) turn() {
	s.currentTurn = (s.currentTurn + 1) % len(s.playerIDs)
}

func (s *shed) buildEndOfTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID: playerID,
		Command:  protocol.Replenish,
		Pile:     s.pile,
	}
}

func legalMoves(pile, toPlay []deck.Card) []int {
	moves := []int{}
	// can play any card on an empty pile
	if len(pile) == 0 {
		for i := range toPlay {
			moves = append(moves, i)
		}
		return moves
	}

	topmostCard := pile[0]
	// two
	if topmostCard.Rank == deck.Two {
		for i := range toPlay {
			moves = append(moves, i)
		}
		return moves
	}

	// seven
	if topmostCard.Rank == deck.Seven {
		for i, tp := range toPlay {
			if wins := sevenBeaters[tp.Rank]; wins {
				moves = append(moves, i)
			}
		}
		return moves
	}

	for i, tp := range toPlay {
		tpValue := cardValues[tp.Rank]
		pileValue := cardValues[pile[0].Rank]

		if tpValue >= pileValue {
			moves = append(moves, i)
		}
	}

	return moves
}

func cardsUnique(cards []deck.Card) bool {
	seen := map[deck.Card]struct{}{}
	for _, c := range cards {
		if _, ok := seen[c]; ok {
			return false
		}
		seen[c] = struct{}{}
	}
	return true
}
