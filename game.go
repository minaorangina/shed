package shed

import (
	"errors"
	"fmt"
	"math/rand"
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

		s.awaitingResponse = true
		return msgs, nil

	case clearDeck:
		s.awaitingResponse = true
		return s.attemptMove(protocol.PlayHand), nil

	case clearCards:
		if len(currentPlayerCards.Hand) > 0 {
			s.awaitingResponse = true
			return s.attemptMove(protocol.PlayHand), nil
		}

		if len(currentPlayerCards.Seen) > 0 {
			s.awaitingResponse = true
			return s.attemptMove(protocol.PlaySeen), nil
		}

		if len(currentPlayerCards.Unseen) > 0 {
			unseenIndices := []int{}
			for i := range currentPlayerCards.Unseen {
				unseenIndices = append(unseenIndices, i)
			}

			toSend := []OutboundMessage{{
				PlayerID:         s.currentPlayerID,
				Command:          protocol.PlayUnseen,
				Hand:             s.playerCards[s.currentPlayerID].Hand,
				Seen:             s.playerCards[s.currentPlayerID].Seen,
				Pile:             s.pile,
				Moves:            unseenIndices,
				Opponents:        buildOpponents(s.currentPlayerID, s.playerCards),
				AwaitingResponse: true,
			}}
			for _, id := range s.playerIDs {
				if id != s.currentPlayerID {
					toSend = append(toSend, OutboundMessage{
						PlayerID:    id,
						Command:     protocol.Turn,
						CurrentTurn: s.currentPlayerID,
						Hand:        s.playerCards[id].Hand,
						Seen:        s.playerCards[id].Seen,
						Pile:        s.pile,
						Opponents:   buildOpponents(s.currentPlayerID, s.playerCards),
					})
				}
			}

			s.awaitingResponse = true
			return toSend, nil
		}

		// this player is already out of the game, handle this situation somehow
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

	if msg.Command == protocol.SkipTurn { // ack
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
			s.completeMove(msg)
			s.pluckFromDeck(msg)

			// return messages with ack from current player expected.
			toSend := []OutboundMessage{{
				PlayerID:         s.currentPlayerID,
				Command:          protocol.ReplenishHand,
				Hand:             s.playerCards[s.currentPlayerID].Hand,
				Seen:             s.playerCards[s.currentPlayerID].Seen,
				Pile:             s.pile,
				AwaitingResponse: true,
			}}

			for _, id := range s.playerIDs {
				if id != s.currentPlayerID {
					toSend = append(toSend, s.buildEndOfTurnMessage(id))
				}
			}

			s.awaitingResponse = true
			return toSend, nil

		case protocol.ReplenishHand:
			// If deck is empty, switch to stage 2
			if len(s.deck) == 0 {
				s.stage = clearCards
			}

			s.awaitingResponse = false
			s.turn()
			return nil, nil
		}
	}

	// stage 2
	if s.stage == clearCards {
		switch msg.Command {

		case protocol.EndOfTurn: // ack
			s.awaitingResponse = false
			s.turn()
			return nil, nil

		case protocol.PlayHand:
			s.completeMove(msg)
			s.pluckFromDeck(msg)

			toSend := s.buildEndOfTurnMessages(protocol.EndOfTurn)
			s.awaitingResponse = true
			return toSend, nil

		case protocol.PlaySeen:
			s.completeMove(msg)

			toSend := s.buildEndOfTurnMessages(protocol.EndOfTurn)
			s.awaitingResponse = true
			return toSend, nil

		case protocol.PlayUnseen:
			if len(msg.Decision) == 0 {
				panic("empty decision")
			}
			// possible optimisation: could precalculate legal Unseen card moves
			cardIdx := msg.Decision[0]
			legalMoves := getLegalMoves(s.pile,
				[]deck.Card{s.playerCards[s.currentPlayerID].Unseen[cardIdx]})

			var cmd protocol.Cmd
			if len(legalMoves) == 0 {
				s.pickUpPile(s.currentPlayerID)
				cmd = protocol.UnseenFailure
			} else {
				s.completeMove(msg)
				cmd = protocol.UnseenSuccess
			}

			toSend := s.buildEndOfTurnMessages(cmd)
			s.awaitingResponse = true
			return toSend, nil
		}
	}

	return nil, errors.New("invalid game state")
}

// step 1 of 2 in a player playing their cards (Hand or Seen)
func (s *shed) attemptMove(currentPlayerCmd protocol.Cmd) []OutboundMessage {
	s.awaitingResponse = true

	var cards []deck.Card
	switch currentPlayerCmd {
	case protocol.PlayHand:
		cards = s.playerCards[s.currentPlayerID].Hand

	case protocol.PlaySeen:
		cards = s.playerCards[s.currentPlayerID].Seen

	default:
		panic(fmt.Sprintf("unrecognised move protocol %s", currentPlayerCmd))
	}

	legalMoves := getLegalMoves(s.pile, cards)
	if len(legalMoves) > 0 {
		toSend := s.buildTurnMessages(currentPlayerCmd, legalMoves)
		return toSend
	}

	// no legal moves
	s.pickUpPile(s.currentPlayerID)

	toSend := s.buildSkipTurnMessages(protocol.SkipTurn)
	return toSend
}

// step 2 of 2 of a player playing their cards (Hand or Seen)
func (s *shed) completeMove(msg InboundMessage) {

	var (
		toPile      = []deck.Card{}
		newCards    *[]deck.Card
		newCardsSet = map[deck.Card]struct{}{}
	)

	switch msg.Command {
	case protocol.PlayHand:
		newCards = &s.playerCards[s.currentPlayerID].Hand
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayerID].Hand)

	case protocol.PlaySeen:
		newCards = &s.playerCards[s.currentPlayerID].Seen
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayerID].Seen)

	case protocol.PlayUnseen:
		newCards = &s.playerCards[s.currentPlayerID].Unseen
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayerID].Unseen)
	}

	for _, cardIdx := range msg.Decision {
		toPile = append(toPile, (*newCards)[cardIdx])
		delete(newCardsSet, (*newCards)[cardIdx])
	}

	s.pile = append(s.pile, toPile...)
	*newCards = setToCardSlice(newCardsSet)
}

func (s *shed) pluckFromDeck(msg InboundMessage) {
	if len(s.deck) == 0 {
		return
	}
	fromDeck := s.deck.Deal(len(msg.Decision))
	s.playerCards[s.currentPlayerID].Hand = append(s.playerCards[s.currentPlayerID].Hand, fromDeck...)
}

func (s *shed) pickUpPile(currentPlayerID string) {
	currentPlayerCards := s.playerCards[currentPlayerID]
	currentPlayerCards.Hand = append(currentPlayerCards.Hand, s.pile...)
	s.pile = []deck.Card{}
}

func (s *shed) turn() {
	s.currentTurnIdx = (s.currentTurnIdx + 1) % len(s.playerIDs)
	s.currentPlayerID = s.playerIDs[s.currentTurnIdx]
}

func (s *shed) buildSkipTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.SkipTurn,
		CurrentTurn: s.currentPlayerID,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
	}
}

func (s *shed) buildSkipTurnMessages(currentPlayerCmd protocol.Cmd) []OutboundMessage {
	currentPlayerMsg := OutboundMessage{
		PlayerID:         s.currentPlayerID,
		Command:          currentPlayerCmd, // always SkipTurn
		Hand:             s.playerCards[s.currentPlayerID].Hand,
		Seen:             s.playerCards[s.currentPlayerID].Seen,
		Pile:             s.pile,
		Opponents:        buildOpponents(s.currentPlayerID, s.playerCards),
		AwaitingResponse: true,
	}

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, id := range s.playerIDs {
		if id != s.currentPlayerID {
			toSend = append(toSend, s.buildSkipTurnMessage(id))
		}
	}

	return toSend
}

func (s *shed) buildTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.Turn,
		CurrentTurn: s.currentPlayerID,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
	}
}

func (s *shed) buildTurnMessages(currentPlayerCmd protocol.Cmd, moves []int) []OutboundMessage {
	currentPlayerMsg := OutboundMessage{
		PlayerID:         s.currentPlayerID,
		Command:          currentPlayerCmd,
		Hand:             s.playerCards[s.currentPlayerID].Hand,
		Seen:             s.playerCards[s.currentPlayerID].Seen,
		Pile:             s.pile,
		Moves:            moves,
		Opponents:        buildOpponents(s.currentPlayerID, s.playerCards),
		AwaitingResponse: true,
	}

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, id := range s.playerIDs {
		if id != s.currentPlayerID {
			toSend = append(toSend, s.buildTurnMessage(id))
		}
	}

	return toSend
}

func (s *shed) buildEndOfTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.EndOfTurn,
		CurrentTurn: s.currentPlayerID,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
	}
}

func (s *shed) buildEndOfTurnMessages(currentPlayerCommand protocol.Cmd) []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, id := range s.playerIDs {
		msg := s.buildEndOfTurnMessage(id)
		if id == s.currentPlayerID {
			msg.Command = currentPlayerCommand
			msg.AwaitingResponse = true
		}
		toSend = append(toSend, msg)
	}

	return toSend
}
