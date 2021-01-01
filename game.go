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
	currentPlayerID  string
	currentTurnIdx   int
	stage            Stage
	awaitingResponse bool
}

type ShedOpts struct {
	deck             deck.Deck
	pile             []deck.Card
	playerCards      map[string]*PlayerCards
	playerIDs        []string
	currentPlayerID  string
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
		currentPlayerID:  opts.currentPlayerID,
		stage:            opts.stage,
		awaitingResponse: opts.awaitingResponse,
	}

	if len(s.playerIDs) > 0 {
		for i, id := range s.playerIDs {
			if id == s.currentPlayerID {
				s.currentTurnIdx = i
				break
			}
		}
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
	s.currentTurnIdx = rand.Intn(len(s.playerIDs) - 1)

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
	currentPlayerCards := s.playerCards[s.currentPlayerID]

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
		if len(s.playerCards[s.currentPlayerID].Hand) > 0 {
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
	if msg.PlayerID != s.currentPlayerID {
		return nil, fmt.Errorf("unexpected message from player %s", msg.PlayerID)
	}

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
			s.pluckFromDeck(msg)

			// return messages with no response expected.
			toSend := []OutboundMessage{{
				PlayerID: playerIDFromMsg,
				Command:  protocol.ReplenishHand,
				Hand:     s.playerCards[playerIDFromMsg].Hand,
				Pile:     s.pile,
			}}

			for _, id := range s.playerIDs {
				if id != s.currentPlayerID {
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
			s.pluckFromDeck(msg)

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
	s.currentTurnIdx = (s.currentTurnIdx + 1) % len(s.playerIDs)
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

		return setToIntSlice(moves)
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
		return setToIntSlice(moves)
	}

	// seven
	if topmostCard.Rank == deck.Seven {
		for i, tp := range toPlay {
			if wins := sevenBeaters[tp.Rank]; wins {
				moves[i] = struct{}{}
			}
		}
		return setToIntSlice(moves)
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

	return setToIntSlice(moves)
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

// step 1 of 2 in a player playing their hand
func (s *shed) attemptToPlayHand() ([]OutboundMessage, error) {
	msgs := []OutboundMessage{}
	playerCards := s.playerCards[s.currentPlayerID]

	legalMoves := getLegalMoves(s.pile, playerCards.Hand)
	if len(legalMoves) > 0 {
		s.awaitingResponse = true

		for _, recipientID := range s.playerIDs {
			message := OutboundMessage{
				PlayerID:    recipientID,
				CurrentTurn: s.currentPlayerID,
				Command:     protocol.Turn,
				Hand:        s.playerCards[recipientID].Hand,
				Seen:        s.playerCards[recipientID].Seen,
				Opponents:   buildOpponents(recipientID, s.playerCards),
			}
			if recipientID == s.currentPlayerID {
				message.Command = protocol.PlayHand
				message.Moves = legalMoves
				message.AwaitingResponse = true
			}

			msgs = append(msgs, message)

		}

		return msgs, nil
	}

	// no legal moves
	playerCards.Hand = append(s.playerCards[s.currentPlayerID].Hand, s.pile...)
	s.pile = []deck.Card{}
	// still want an ack from the player
	s.awaitingResponse = true

	for _, recipientID := range s.playerIDs {
		message := OutboundMessage{
			PlayerID:    s.currentPlayerID,
			Command:     protocol.Turn,
			CurrentTurn: s.currentPlayerID,
			Hand:        s.playerCards[s.currentPlayerID].Hand,
			Seen:        s.playerCards[s.currentPlayerID].Seen,
		}
		if recipientID == s.currentPlayerID {
			message.Command = protocol.NoLegalMoves
			message.AwaitingResponse = true
		}

		msgs = append(msgs, message)
	}

	return msgs, nil
}

// step 2 of 2 of a player playing their hand
func (s *shed) playHand(msg InboundMessage) {
	toPile := []deck.Card{}
	newHand := cardSliceToSet(s.playerCards[s.currentPlayerID].Hand)

	for _, cardIdx := range msg.Decision {
		toPile = append(toPile, s.playerCards[s.currentPlayerID].Hand[cardIdx])
		delete(newHand, s.playerCards[s.currentPlayerID].Hand[cardIdx])
	}

	s.pile = append(s.pile, toPile...)
	s.playerCards[s.currentPlayerID].Hand = setToCardSlice(newHand)
}

func (s *shed) pluckFromDeck(msg InboundMessage) {
	if len(s.deck) == 0 {
		return
	}
	fromDeck := s.deck.Deal(len(msg.Decision))
	s.playerCards[s.currentPlayerID].Hand = append(s.playerCards[s.currentPlayerID].Hand, fromDeck...)
}

func setToIntSlice(set map[int]struct{}) []int {
	s := []int{}
	for key := range set {
		s = append(s, key)
	}

	sort.Ints(s)

	return s
}

func setToCardSlice(set map[deck.Card]struct{}) []deck.Card {
	s := []deck.Card{}
	for key := range set {
		s = append(s, key)
	}
	return s
}

func cardSliceToSet(s []deck.Card) map[deck.Card]struct{} {
	set := map[deck.Card]struct{}{}
	for _, key := range s {
		set[key] = struct{}{}
	}
	return set
}
