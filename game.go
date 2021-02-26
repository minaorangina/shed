package shed

import (
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
	reorgSeenOffset = 3
	numCardsInGroup = 3
)

type PlayerCards struct {
	Hand, Seen, Unseen []deck.Card
	UnseenVisibility   map[deck.Card]bool
}

func NewPlayerCards(
	hand, seen, unseen []deck.Card,
	unseenVisibility map[deck.Card]bool,
) *PlayerCards {
	if hand == nil {
		hand = []deck.Card{}
	}
	if seen == nil {
		seen = []deck.Card{}
	}
	if unseen == nil {
		unseen = []deck.Card{}
	}
	if unseenVisibility == nil {
		unseenVisibility = map[deck.Card]bool{}
		for _, c := range unseen {
			unseenVisibility[c] = false
		}
	}

	return &PlayerCards{
		Hand:             hand,
		Seen:             seen,
		Unseen:           unseen,
		UnseenVisibility: unseenVisibility,
	}
}

type Game interface {
	Start(playerInfo []PlayerInfo) error
	Next() ([]OutboundMessage, error)
	ReceiveResponse([]InboundMessage) ([]OutboundMessage, error)
	AwaitingResponse() protocol.Cmd
}

type shed struct {
	Deck            deck.Deck
	Pile            []deck.Card
	PlayerCards     map[string]*PlayerCards
	PlayerInfo      []PlayerInfo
	ActivePlayers   []PlayerInfo
	FinishedPlayers []PlayerInfo
	CurrentPlayer   PlayerInfo
	CurrentTurnIdx  int
	Stage           Stage
	ExpectedCommand protocol.Cmd
	GameOver        bool
	unseenDecision  *InboundMessage
}

type ShedOpts struct {
	Deck            deck.Deck
	Pile            []deck.Card
	PlayerCards     map[string]*PlayerCards
	PlayerInfo      []PlayerInfo
	FinishedPlayers []PlayerInfo
	CurrentPlayer   PlayerInfo
	Stage           Stage
	ExpectedCommand protocol.Cmd
}

// NewShed constructs a new game of Shed
func NewShed(opts ShedOpts) *shed {
	s := &shed{
		Deck:            opts.Deck,
		Pile:            opts.Pile,
		PlayerCards:     opts.PlayerCards,
		PlayerInfo:      opts.PlayerInfo,
		FinishedPlayers: opts.FinishedPlayers,
		CurrentPlayer:   opts.CurrentPlayer,
		Stage:           opts.Stage,
		ExpectedCommand: opts.ExpectedCommand,
	}

	if len(s.PlayerInfo) > 0 {
		for i, info := range s.PlayerInfo {
			if info.PlayerID == s.CurrentPlayer.PlayerID {
				s.CurrentTurnIdx = i
				break
			}
		}
	}

	if s.Deck == nil {
		s.Deck = deck.New()
		s.Deck.Shuffle()
	}
	if s.Pile == nil {
		s.Pile = []deck.Card{}
	}
	if s.PlayerCards == nil {
		s.PlayerCards = map[string]*PlayerCards{}
	}
	if s.PlayerInfo == nil {
		// new game
		s.PlayerInfo = []PlayerInfo{}
		s.ActivePlayers = []PlayerInfo{}
		s.FinishedPlayers = []PlayerInfo{}
	} else if s.FinishedPlayers == nil {
		s.PlayerInfo = opts.PlayerInfo
		s.ActivePlayers = opts.PlayerInfo
	} else {
		// work out who is still playing the game
		stillPlaying := []PlayerInfo{}
		for _, pi := range opts.PlayerInfo {
			for _, fp := range opts.FinishedPlayers {
				if fp.PlayerID != pi.PlayerID {
					stillPlaying = append(stillPlaying, pi)
				}
			}
		}
		s.PlayerInfo = opts.PlayerInfo
		s.ActivePlayers = stillPlaying
	}

	return s
}

func (s *shed) AwaitingResponse() protocol.Cmd {
	return s.ExpectedCommand
}

func (s *shed) Start(playerInfo []PlayerInfo) error {
	if s == nil {
		return ErrNilGame
	}
	if len(playerInfo) < minPlayers {
		return ErrTooFewPlayers
	}
	if len(playerInfo) > maxPlayers {
		return ErrTooManyPlayers
	}

	s.PlayerInfo = playerInfo
	s.ActivePlayers = s.PlayerInfo

	for _, info := range playerInfo {
		playerCards := NewPlayerCards(
			s.Deck.Deal(numCardsInGroup),
			s.Deck.Deal(numCardsInGroup),
			s.Deck.Deal(numCardsInGroup),
			nil,
		)
		s.PlayerCards[info.PlayerID] = playerCards
	}

	rand.Seed(time.Now().UnixNano())
	s.CurrentTurnIdx = rand.Intn(len(s.PlayerInfo) - 1)
	s.CurrentPlayer = s.PlayerInfo[s.CurrentTurnIdx]

	return nil
}

func (s *shed) Next() ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.PlayerCards == nil || len(s.PlayerCards) == 0 {
		return nil, ErrNoPlayers
	}
	if s.ExpectedCommand != protocol.Null {
		return nil, ErrGameAwaitingResponse
	}
	if s.GameOver {
		return s.buildGameOverMessages(), nil
	}

	currentPlayerCards := s.PlayerCards[s.CurrentPlayer.PlayerID]

	switch s.Stage {
	case preGame:
		msgs := s.buildReorgMessages()
		s.ExpectedCommand = protocol.Reorg
		return msgs, nil

	case clearDeck:
		msgs, legalMoves := s.attemptMove(protocol.PlayHand)
		if legalMoves {
			s.ExpectedCommand = protocol.PlayHand
		} else {
			s.ExpectedCommand = protocol.SkipTurn
		}
		return msgs, nil

	case clearCards:
		if len(currentPlayerCards.Hand) > 0 {
			msgs, legalMoves := s.attemptMove(protocol.PlayHand)
			if legalMoves {
				s.ExpectedCommand = protocol.PlayHand
			} else {
				s.ExpectedCommand = protocol.SkipTurn
			}
			return msgs, nil
		}

		if len(currentPlayerCards.Seen) > 0 {
			msgs, legalMoves := s.attemptMove(protocol.PlaySeen)
			if legalMoves {
				s.ExpectedCommand = protocol.PlaySeen
			} else {
				s.ExpectedCommand = protocol.SkipTurn
			}
			return msgs, nil
		}

		if len(currentPlayerCards.Unseen) > 0 {
			unseenIndices := []int{}
			for i := range currentPlayerCards.Unseen {
				unseenIndices = append(unseenIndices, i)
			}

			toSend := s.buildTurnMessages(protocol.PlayUnseen, unseenIndices)
			s.ExpectedCommand = protocol.PlayUnseen
			return toSend, nil
		}

		// this player has finished!
		return nil, nil
	}

	// this shouldn't happen
	return nil, fmt.Errorf("could not match game stage %d", s.Stage)
}

func (s *shed) ReceiveResponse(inboundMsgs []InboundMessage) ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.PlayerCards == nil {
		return nil, ErrNoPlayers
	}
	if s.ExpectedCommand == protocol.Null {
		return nil, ErrGameUnexpectedResponse
	}
	if s.GameOver {
		return s.buildGameOverMessages(), nil
	}

	// stage 0
	if s.Stage == preGame {
		numPlayers, numMessages := len(s.PlayerInfo), len(inboundMsgs)
		if numPlayers != numMessages {
			return nil, fmt.Errorf("expected %d messages, got %d", numMessages, numPlayers)
		}

		for _, m := range inboundMsgs {
			cardIndicesSet := intSliceToSet([]int{0, 1, 2, 3, 4, 5})
			newHand, newSeen := []deck.Card{}, []deck.Card{}

			for _, v := range m.Decision {
				newHand = append(newHand, s.getReorgCard(m.PlayerID, v))
				delete(cardIndicesSet, v)
			}

			cardIndices := setToIntSlice(cardIndicesSet)
			sort.Ints(cardIndices)

			for _, v := range cardIndices {
				newSeen = append(newSeen, s.getReorgCard(m.PlayerID, v))
			}

			s.PlayerCards[m.PlayerID].Hand = newHand
			s.PlayerCards[m.PlayerID].Seen = newSeen
		}

		// switch to stage 1
		s.Stage = clearDeck
		s.ExpectedCommand = protocol.Null
		return nil, nil
	}

	msg := inboundMsgs[0]
	if msg.PlayerID != s.CurrentPlayer.PlayerID {
		err := fmt.Errorf("unexpected message from player %s", msg.PlayerID)
		return []OutboundMessage{s.buildErrorMessage(msg.PlayerID, err)}, err
	}
	if msg.Command != s.ExpectedCommand {
		err := fmt.Errorf("unexpected command - got %s, want %s", msg.Command.String(), s.ExpectedCommand.String())
		return []OutboundMessage{s.buildErrorMessage(s.CurrentPlayer.PlayerID, err)}, err
	}

	if msg.Command == protocol.Burn { // ack
		// player gets another turn
		s.ExpectedCommand = protocol.Null
		return nil, nil
	}

	if msg.Command == protocol.SkipTurn { // ack
		s.ExpectedCommand = protocol.Null
		s.turn()
		return nil, nil
	}

	// stage 1
	if s.Stage == clearDeck {
		if len(inboundMsgs) != 1 {
			return nil, fmt.Errorf("expected one message, got %d", len(inboundMsgs))
		}

		switch msg.Command {

		case protocol.PlayHand:
			// check this is a legal move. this has already been done, but worth
			// double checking in case of client tampering.

			// If playing more than one card, they must be of the same rank
			if len(msg.Decision) > 1 {
				pc := s.PlayerCards[s.CurrentPlayer.PlayerID].Hand
				referenceCard := pc[msg.Decision[0]]

				for _, idx := range msg.Decision[1:] {
					if pc[idx].Rank != referenceCard.Rank {
						return s.buildErrorMessages(ErrInvalidMove), ErrInvalidMove
					}
				}
			}

			s.completeMove(msg)

			// Pick up a new card if necessary
			if len(s.PlayerCards[s.CurrentPlayer.PlayerID].Hand) < numCardsInGroup {
				s.pluckFromDeck(msg)
			}

			if isBurn(s.Pile) {
				// Maybe in future the old cards are banished out of sight but not deleted
				// Useful for undo mechanism etc
				s.Pile = []deck.Card{}
				s.ExpectedCommand = protocol.Burn
				return s.buildBurnMessages(), nil
			}

			msgs := s.buildReplenishHandMessages()
			s.ExpectedCommand = protocol.ReplenishHand
			return msgs, nil

		case protocol.ReplenishHand:
			// If deck is empty, switch to stage 2
			if len(s.Deck) == 0 {
				s.Stage = clearCards
			}

			s.ExpectedCommand = protocol.Null
			s.turn()
			return nil, nil
		}
	}

	// stage 2
	if s.Stage == clearCards {
		switch msg.Command {

		case protocol.PlayerFinished: // ack
			s.ExpectedCommand = protocol.Null
			s.moveToFinishedPlayers() // handles the next turn

			if s.gameIsOver() {
				s.GameOver = true
				// move the remaining player
				s.moveToFinishedPlayers()
				return s.buildGameOverMessages(), nil
			}

			return nil, nil

		case protocol.EndOfTurn: // ack
			s.ExpectedCommand = protocol.Null
			s.turn()
			return nil, nil

		case protocol.UnseenSuccess: // ack
			s.completeMove(*s.unseenDecision)
			s.unseenDecision = nil

			if s.playerHasFinished() {
				s.ExpectedCommand = protocol.PlayerFinished
				return s.buildPlayerFinishedMessages(), nil
			}

			s.ExpectedCommand = protocol.Null
			s.turn()
			return nil, nil

		case protocol.UnseenFailure: // ack
			// Must play the card to pick up the card
			s.completeMove(*s.unseenDecision)
			s.unseenDecision = nil

			s.pickUpPile()

			s.ExpectedCommand = protocol.Null
			s.turn()
			return nil, nil

		case protocol.PlayHand, protocol.PlaySeen:
			s.completeMove(msg)

			if isBurn(s.Pile) {
				// Maybe in future the old cards are banished out of sight but not deleted
				// Useful for undo mechanism etc
				s.Pile = []deck.Card{}
				s.ExpectedCommand = protocol.Burn
				return s.buildBurnMessages(), nil
			}

			if s.playerHasFinished() {
				s.ExpectedCommand = protocol.PlayerFinished
				return s.buildPlayerFinishedMessages(), nil
			}

			toSend := s.buildEndOfTurnMessages(protocol.EndOfTurn)
			s.ExpectedCommand = protocol.EndOfTurn
			return toSend, nil

		case protocol.PlayUnseen:
			if len(msg.Decision) != 1 {
				return []OutboundMessage{
					s.buildErrorMessage(s.CurrentPlayer.PlayerID, ErrPlayOneCard),
				}, ErrPlayOneCard
			}
			// possible optimisation: could precalculate legal Unseen card moves

			// The player plays their chosen card regardless of the legality of the move
			// If it's a legal move, then this is fine.
			// If it's not a legal move, the player will pick up the pile anyway.
			s.unseenDecision = &msg

			// Flip chosen card
			cardIdx := msg.Decision[0]
			chosenCard := s.PlayerCards[s.CurrentPlayer.PlayerID].Unseen[cardIdx]
			s.PlayerCards[s.CurrentPlayer.PlayerID].UnseenVisibility[chosenCard] = true

			legalMoves := getLegalMoves(s.Pile, []deck.Card{chosenCard})

			if len(legalMoves) > 0 {
				s.ExpectedCommand = protocol.UnseenSuccess
				return s.buildEndOfTurnMessages(protocol.UnseenSuccess), nil
			}

			s.ExpectedCommand = protocol.UnseenFailure
			return s.buildEndOfTurnMessages(protocol.UnseenFailure), nil

		}
	}

	return nil, ErrInvalidGameState
}

// step 1 of 2 in a player playing their cards (Hand or Seen)
func (s *shed) attemptMove(currentPlayerCmd protocol.Cmd) ([]OutboundMessage, bool) {
	var cards []deck.Card
	switch currentPlayerCmd {

	case protocol.PlayHand:
		cards = s.PlayerCards[s.CurrentPlayer.PlayerID].Hand

	case protocol.PlaySeen:
		cards = s.PlayerCards[s.CurrentPlayer.PlayerID].Seen

	default:
		panic(fmt.Sprintf("unrecognised move protocol %s", currentPlayerCmd))
	}

	legalMoves := getLegalMoves(s.Pile, cards)
	if len(legalMoves) > 0 {
		toSend := s.buildTurnMessages(currentPlayerCmd, legalMoves)
		return toSend, true
	}

	// no legal moves
	s.pickUpPile()

	toSend := s.buildSkipTurnMessages(protocol.SkipTurn)
	return toSend, false
}

// step 2 of 2 of a player playing their cards (Hand or Seen)
func (s *shed) completeMove(msg InboundMessage) {

	var (
		toPile       = []deck.Card{}
		cardGroup    *[]deck.Card
		cardGroupSet = map[deck.Card]struct{}{}
	)

	switch msg.Command {
	case protocol.PlayHand:
		cardGroup = &s.PlayerCards[s.CurrentPlayer.PlayerID].Hand
		cardGroupSet = cardSliceToSet(s.PlayerCards[s.CurrentPlayer.PlayerID].Hand)

	case protocol.PlaySeen:
		cardGroup = &s.PlayerCards[s.CurrentPlayer.PlayerID].Seen
		cardGroupSet = cardSliceToSet(s.PlayerCards[s.CurrentPlayer.PlayerID].Seen)

	case protocol.PlayUnseen:
		cardGroup = &s.PlayerCards[s.CurrentPlayer.PlayerID].Unseen
		cardGroupSet = cardSliceToSet(s.PlayerCards[s.CurrentPlayer.PlayerID].Unseen)
	}

	for _, cardIdx := range msg.Decision {
		toPile = append(toPile, (*cardGroup)[cardIdx])
		delete(cardGroupSet, (*cardGroup)[cardIdx])
	}

	s.Pile = append(s.Pile, toPile...)
	*cardGroup = setToCardSlice(cardGroupSet)
}

func (s *shed) pluckFromDeck(msg InboundMessage) {
	if len(s.Deck) == 0 {
		return
	}
	fromDeck := s.Deck.Deal(len(msg.Decision))
	s.PlayerCards[s.CurrentPlayer.PlayerID].Hand = append(s.PlayerCards[s.CurrentPlayer.PlayerID].Hand, fromDeck...)
}

func (s *shed) pickUpPile() {
	playerID := s.CurrentPlayer.PlayerID
	currentPlayerCards := s.PlayerCards[playerID]
	currentPlayerCards.Hand = append(currentPlayerCards.Hand, s.Pile...)
	s.Pile = []deck.Card{}
}

func (s *shed) playerHasFinished() bool {
	pc := s.PlayerCards[s.CurrentPlayer.PlayerID]
	return len(pc.Hand) == 0 &&
		len(pc.Seen) == 0 &&
		len(pc.Unseen) == 0
}

func (s *shed) gameIsOver() bool {
	return len(s.ActivePlayers) == 1
}

func (s *shed) turn() {
	s.CurrentTurnIdx = (s.CurrentTurnIdx + 1) % len(s.ActivePlayers)
	s.CurrentPlayer = s.ActivePlayers[s.CurrentTurnIdx]
}

func (s *shed) moveToFinishedPlayers() {
	if len(s.ActivePlayers) == 1 {
		s.FinishedPlayers = append(s.FinishedPlayers, s.ActivePlayers[0])
		s.ActivePlayers = []PlayerInfo{}
		return
	}

	s.ActivePlayers = append(s.ActivePlayers[:s.CurrentTurnIdx],
		s.ActivePlayers[(s.CurrentTurnIdx+1)%len(s.ActivePlayers):]...)

	s.FinishedPlayers = append(s.FinishedPlayers, s.CurrentPlayer)

	s.CurrentPlayer = s.ActivePlayers[s.CurrentTurnIdx]
}

func (s *shed) getReorgCard(playerID string, choice int) deck.Card {
	oldHand := s.PlayerCards[playerID].Hand
	oldSeen := s.PlayerCards[playerID].Seen

	var card deck.Card
	if choice > 2 {
		card = oldSeen[choice-reorgSeenOffset]
	} else {
		card = oldHand[choice]
	}

	return card
}
