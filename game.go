package shed

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

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
		deck:             opts.deck,
		pile:             opts.pile,
		playerCards:      opts.playerCards,
		playerIDs:        opts.playerIDs,
		currentTurn:      opts.currentTurn,
		stage:            opts.stage,
		awaitingResponse: opts.awaitingResponse,
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

	s.playerIDs = playerIDs

	s.deck.Shuffle()

	for _, id := range playerIDs {
		playerCards := &PlayerCards{
			Hand:   s.deck.Deal(3),
			Seen:   s.deck.Deal(3),
			Unseen: s.deck.Deal(3),
		}
		s.playerCards[id] = playerCards
	}

	rand.Seed(time.Now().UnixNano())
	s.currentTurn = rand.Intn(len(s.playerIDs) - 1)

	return nil
}

func (s *shed) Next() ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.playerCards == nil || len(s.playerCards) == 0 {
		return nil, ErrNoPlayers
	}
	if s.awaitingResponse {
		return nil, ErrGameAwaitingResponse
	}

	msgs := []OutboundMessage{}

	switch s.stage {
	case preGame:
		s.awaitingResponse = true

		for _, id := range s.playerIDs {
			m := OutboundMessage{
				PlayerID:         id,
				Command:          protocol.Reorg,
				Hand:             s.playerCards[id].Hand,
				Seen:             s.playerCards[id].Seen,
				AwaitingResponse: true,
			}
			msgs = append(msgs, m)
		}

		return msgs, nil

	case clearDeck:
		return s.attemptToPlayHand()

	case clearCards:
		currentPlayer := s.playerIDs[s.currentTurn]
		if len(s.playerCards[currentPlayer].Hand) > 0 {
			return s.attemptToPlayHand()
		}
		return nil, nil
	}

	// this shouldn't happen
	return nil, fmt.Errorf("could not match game stage %d", s.stage)
}

func (s *shed) ReceiveResponse(inboundMsgs []InboundMessage) ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.playerCards == nil {
		return nil, ErrNoPlayers
	}
	if !s.awaitingResponse {
		return nil, ErrGameUnexpectedResponse
	}

	// stage 0
	if s.stage == preGame {
		numPlayers, numMessages := len(s.playerIDs), len(inboundMsgs)
		if numPlayers != numMessages {
			return nil, fmt.Errorf("expected %d messages, got %d", numMessages, numPlayers)
		}

		for _, m := range inboundMsgs {
			s.playerCards[m.PlayerID].Hand = m.Hand
			s.playerCards[m.PlayerID].Seen = m.Seen
		}

		// switch to stage 1
		s.stage = clearDeck
		s.awaitingResponse = false
		return nil, nil
	}

	msg := inboundMsgs[0]
	playerID := msg.PlayerID // check it's an id we recognise (gameengine responsibility?)

	if msg.Command == protocol.NoLegalMoves { // ack
		s.awaitingResponse = false
		s.turn()
		return nil, nil
	}

	// stage 1
	if s.stage == clearDeck {
		if len(inboundMsgs) != 1 {
			return nil, fmt.Errorf("expected one message, got %d", len(inboundMsgs))
		}

		switch msg.Command {

		case protocol.PlayHand:
			// check this is a legal move. this has already been done, but worth
			// double checking in case of client tampering.

			s.playHand(msg)
			ok := s.pluckFromDeck(msg)
			_ = ok

			// return messages with no response expected.
			toSend := []OutboundMessage{{
				PlayerID: playerID,
				Command:  protocol.ReplenishHand,
				Hand:     s.playerCards[playerID].Hand,
				Pile:     s.pile,
			}}

			for _, id := range s.playerIDs {
				if id != playerID {
					toSend = append(toSend, s.buildEndOfTurnMessage(id))
				}
			}

			// If deck is empty, switch to stage 2
			if len(s.deck) == 0 {
				s.stage = clearCards
			}

			s.awaitingResponse = false
			s.turn()
			return toSend, nil
		}

	}

	// stage 2
	if s.stage == clearCards {
		switch msg.Command {
		case protocol.PlayHand:

			s.playHand(msg)
			ok := s.pluckFromDeck(msg)
			_ = ok

			toSend := []OutboundMessage{}
			for _, id := range s.playerIDs {
				toSend = append(toSend, s.buildEndOfTurnMessage(id))
			}

			s.awaitingResponse = false
			s.turn()
			return toSend, nil
		}
	}

	return nil, errors.New("invalid game state")
}

func (s *shed) turn() {
	s.currentTurn = (s.currentTurn + 1) % len(s.playerIDs)
}

func (s *shed) buildEndOfTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID: playerID,
		Command:  protocol.ReplenishHand,
		Pile:     s.pile,
	}
}

func getLegalMoves(pile, toPlay []deck.Card) []int {
	candidates := []deck.Card{}
	for _, c := range pile {
		if c.Rank != deck.Three {
			candidates = append(candidates, c)
		}
	}

	moves := map[int]struct{}{}

	// can play any card on an empty pile
	if len(candidates) == 0 {
		for i := range toPlay {
			moves[i] = struct{}{}
		}

		return setToSlice(moves)
	}

	// tens (and twos and threes) beat anything
	for i, tp := range toPlay {
		if tp.Rank == deck.Ten {
			moves[i] = struct{}{}
			continue
		}
	}

	topmostCard := candidates[0]
	// two
	if topmostCard.Rank == deck.Two {
		for i := range toPlay {
			moves[i] = struct{}{}
		}
		return setToSlice(moves)
	}

	// seven
	if topmostCard.Rank == deck.Seven {
		for i, tp := range toPlay {
			if wins := sevenBeaters[tp.Rank]; wins {
				moves[i] = struct{}{}
			}
		}
		return setToSlice(moves)
	}

	for i, tp := range toPlay {
		// skip tens (and twos and threes)
		if tp.Rank == deck.Ten {
			continue
		}

		tpValue := cardValues[tp.Rank]
		pileValue := cardValues[topmostCard.Rank]

		if tpValue >= pileValue {
			moves[i] = struct{}{}
		}
	}

	return setToSlice(moves)
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

func setToSlice(set map[int]struct{}) []int {
	s := []int{}
	for key := range set {
		s = append(s, key)
	}

	sort.Ints(s)

	return s
}

// step 1 of 2 in a player playing their hand
func (s *shed) attemptToPlayHand() ([]OutboundMessage, error) {
	msgs := []OutboundMessage{}

	playerID := s.playerIDs[s.currentTurn]
	playerCards := s.playerCards[playerID]

	legalMoves := getLegalMoves(s.pile, playerCards.Hand)
	if len(legalMoves) > 0 {
		s.awaitingResponse = true

		for _, recipientID := range s.playerIDs {
			message := OutboundMessage{
				PlayerID:    recipientID,
				CurrentTurn: playerID,
				Command:     protocol.Turn,
				Hand:        s.playerCards[recipientID].Hand,
				Seen:        s.playerCards[recipientID].Seen,
				Opponents:   buildOpponents(recipientID, s.playerCards),
			}
			if recipientID == playerID {
				message.Command = protocol.PlayHand
				message.Moves = legalMoves
				message.AwaitingResponse = true
			}

			msgs = append(msgs, message)

		}

		return msgs, nil
	}

	// no legal moves
	playerCards.Hand = append(s.playerCards[playerID].Hand, s.pile...)
	s.pile = []deck.Card{}
	// still want an ack from the player
	s.awaitingResponse = true

	for _, recipientID := range s.playerIDs {
		message := OutboundMessage{
			PlayerID:    playerID,
			Command:     protocol.Turn,
			CurrentTurn: playerID,
			Hand:        s.playerCards[playerID].Hand,
			Seen:        s.playerCards[playerID].Seen,
		}
		if recipientID == playerID {
			message.Command = protocol.NoLegalMoves
			message.AwaitingResponse = true
		}

		msgs = append(msgs, message)
	}

	return msgs, nil
}

// step 2 of 2 of a player playing their hand
func (s *shed) playHand(msg InboundMessage) {
	playerID := msg.PlayerID
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
}

func (s *shed) pluckFromDeck(msg InboundMessage) bool {
	if len(s.deck) == 0 {
		return false
	}

	playerID := msg.PlayerID
	// pluck from deck
	fromDeck := s.deck.Deal(len(msg.Decision))
	s.playerCards[playerID].Hand = append(s.playerCards[playerID].Hand, fromDeck...)

	return true
}
