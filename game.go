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
	reorgSeenOffset = 3
	numCardsInGroup = 3
)

type PlayerCards struct {
	Hand, Seen, Unseen []deck.Card
}

type Game interface {
	Start(playerInfo []PlayerInfo) error
	Next() ([]OutboundMessage, error)
	ReceiveResponse([]InboundMessage) ([]OutboundMessage, error)
	AwaitingResponse() protocol.Cmd
}

type shed struct {
	deck             deck.Deck
	pile             []deck.Card
	playerCards      map[string]*PlayerCards
	playerInfo       []PlayerInfo
	activePlayers    []PlayerInfo
	finishedPlayers  []PlayerInfo
	currentPlayer    PlayerInfo
	currentTurnIdx   int
	stage            Stage
	awaitingResponse protocol.Cmd
	gameOver         bool
}

type ShedOpts struct {
	deck             deck.Deck
	pile             []deck.Card
	playerCards      map[string]*PlayerCards
	playerInfo       []PlayerInfo
	finishedPlayers  []PlayerInfo
	currentPlayer    PlayerInfo
	stage            Stage
	awaitingResponse protocol.Cmd
}

// NewShed constructs a new game of Shed
func NewShed(opts ShedOpts) *shed {
	s := &shed{
		deck:             opts.deck,
		pile:             opts.pile,
		playerCards:      opts.playerCards,
		playerInfo:       opts.playerInfo,
		finishedPlayers:  opts.finishedPlayers,
		currentPlayer:    opts.currentPlayer,
		stage:            opts.stage,
		awaitingResponse: opts.awaitingResponse,
	}

	if len(s.playerInfo) > 0 {
		for i, info := range s.playerInfo {
			if info.PlayerID == s.currentPlayer.PlayerID {
				s.currentTurnIdx = i
				break
			}
		}
	}

	if s.deck == nil {
		s.deck = deck.New()
		s.deck.Shuffle()
	}
	if s.pile == nil {
		s.pile = []deck.Card{}
	}
	if s.playerCards == nil {
		s.playerCards = map[string]*PlayerCards{}
	}
	if s.playerInfo == nil {
		// new game
		s.playerInfo = []PlayerInfo{}
		s.activePlayers = []PlayerInfo{}
		s.finishedPlayers = []PlayerInfo{}
	} else if s.finishedPlayers == nil {
		s.playerInfo = opts.playerInfo
		s.activePlayers = opts.playerInfo
	} else {
		// work out who is still playing the game
		stillPlaying := []PlayerInfo{}
		for _, pi := range opts.playerInfo {
			for _, fp := range opts.finishedPlayers {
				if fp.PlayerID != pi.PlayerID {
					stillPlaying = append(stillPlaying, pi)
				}
			}
		}
		s.playerInfo = opts.playerInfo
		s.activePlayers = stillPlaying
	}

	return s
}

func (s *shed) AwaitingResponse() protocol.Cmd {
	return s.awaitingResponse
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

	s.playerInfo = playerInfo
	s.activePlayers = s.playerInfo

	for _, info := range playerInfo {
		playerCards := &PlayerCards{
			Hand:   s.deck.Deal(numCardsInGroup),
			Seen:   s.deck.Deal(numCardsInGroup),
			Unseen: s.deck.Deal(numCardsInGroup),
		}
		s.playerCards[info.PlayerID] = playerCards
	}

	rand.Seed(time.Now().UnixNano())
	s.currentTurnIdx = rand.Intn(len(s.playerInfo) - 1)
	s.currentPlayer = s.playerInfo[s.currentTurnIdx]

	return nil
}

func (s *shed) Next() ([]OutboundMessage, error) {
	if s == nil {
		return nil, ErrNilGame
	}
	if s.playerCards == nil || len(s.playerCards) == 0 {
		return nil, ErrNoPlayers
	}
	if s.awaitingResponse != protocol.Null {
		return nil, ErrGameAwaitingResponse
	}
	if s.gameOver {
		return s.buildGameOverMessages(), nil
	}

	msgs := []OutboundMessage{}
	currentPlayerCards := s.playerCards[s.currentPlayer.PlayerID]

	switch s.stage {
	case preGame:
		for _, info := range s.playerInfo {
			m := OutboundMessage{
				PlayerID:      info.PlayerID,
				Command:       protocol.Reorg,
				Hand:          s.playerCards[info.PlayerID].Hand,
				Seen:          s.playerCards[info.PlayerID].Seen,
				ShouldRespond: true,
			}
			msgs = append(msgs, m)
		}

		s.awaitingResponse = protocol.Reorg
		return msgs, nil

	case clearDeck:
		msgs, legalMoves := s.attemptMove(protocol.PlayHand)
		if legalMoves {
			s.awaitingResponse = protocol.PlayHand
		} else {
			s.awaitingResponse = protocol.SkipTurn
		}
		return msgs, nil

	case clearCards:
		if len(currentPlayerCards.Hand) > 0 {
			msgs, legalMoves := s.attemptMove(protocol.PlayHand)
			if legalMoves {
				s.awaitingResponse = protocol.PlayHand
			} else {
				s.awaitingResponse = protocol.SkipTurn
			}
			return msgs, nil
		}

		if len(currentPlayerCards.Seen) > 0 {
			msgs, legalMoves := s.attemptMove(protocol.PlaySeen)
			if legalMoves {
				s.awaitingResponse = protocol.PlaySeen
			} else {
				s.awaitingResponse = protocol.SkipTurn
			}
			return msgs, nil
		}

		if len(currentPlayerCards.Unseen) > 0 {
			unseenIndices := []int{}
			for i := range currentPlayerCards.Unseen {
				unseenIndices = append(unseenIndices, i)
			}

			toSend := s.buildTurnMessages(protocol.PlayUnseen, unseenIndices)
			s.awaitingResponse = protocol.PlayUnseen
			return toSend, nil
		}

		// this player has finished!
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
	if s.awaitingResponse == protocol.Null {
		return nil, ErrGameUnexpectedResponse
	}
	if s.gameOver {
		return s.buildGameOverMessages(), nil
	}

	// stage 0
	if s.stage == preGame {
		numPlayers, numMessages := len(s.playerInfo), len(inboundMsgs)
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

			s.playerCards[m.PlayerID].Hand = newHand
			s.playerCards[m.PlayerID].Seen = newSeen
		}

		// switch to stage 1
		s.stage = clearDeck
		s.awaitingResponse = protocol.Null
		return nil, nil
	}

	msg := inboundMsgs[0]
	if msg.Command != s.awaitingResponse {
		return nil, fmt.Errorf("unexpected command - got %s, want %s", msg.Command.String(), s.awaitingResponse.String())
	}
	if msg.PlayerID != s.currentPlayer.PlayerID {
		return nil, fmt.Errorf("unexpected message from player %s", msg.PlayerID)
	}

	if msg.Command == protocol.Burn { // ack
		// player gets another turn
		s.awaitingResponse = protocol.Null
		return nil, nil
	}

	if msg.Command == protocol.SkipTurn { // ack
		s.awaitingResponse = protocol.Null
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

			if isBurn(s.pile) {
				s.awaitingResponse = protocol.Burn
				return s.buildBurnMessages(), nil
			}

			toSend := []OutboundMessage{{
				PlayerID:      s.currentPlayer.PlayerID,
				Command:       protocol.ReplenishHand,
				Hand:          s.playerCards[s.currentPlayer.PlayerID].Hand,
				Seen:          s.playerCards[s.currentPlayer.PlayerID].Seen,
				Pile:          s.pile,
				ShouldRespond: true,
			}}

			for _, info := range s.playerInfo {
				if info.PlayerID != s.currentPlayer.PlayerID {
					toSend = append(toSend, s.buildEndOfTurnMessage(info.PlayerID))
				}
			}

			s.awaitingResponse = protocol.ReplenishHand
			return toSend, nil

		case protocol.ReplenishHand:
			// If deck is empty, switch to stage 2
			if len(s.deck) == 0 {
				s.stage = clearCards
			}

			s.awaitingResponse = protocol.Null
			s.turn()
			return nil, nil
		}
	}

	// stage 2
	if s.stage == clearCards {
		switch msg.Command {

		case protocol.PlayerFinished:
			s.awaitingResponse = protocol.Null
			s.moveToFinishedPlayers() // handles the next turn

			if s.gameIsOver() {
				s.gameOver = true
				// move the remaining player
				s.moveToFinishedPlayers()
				return s.buildGameOverMessages(), nil
			}

			return nil, nil

		case protocol.EndOfTurn,
			protocol.UnseenSuccess,
			protocol.UnseenFailure: // ack
			s.awaitingResponse = protocol.Null
			s.turn()
			return nil, nil

		case protocol.PlayHand, protocol.PlaySeen:
			s.completeMove(msg)

			if s.playerHasFinished() {
				s.awaitingResponse = protocol.PlayerFinished
				return s.buildPlayerFinishedMessages(), nil
			}

			toSend := s.buildEndOfTurnMessages(protocol.EndOfTurn)
			s.awaitingResponse = protocol.EndOfTurn
			return toSend, nil

		case protocol.PlayUnseen:
			if len(msg.Decision) != 1 {
				return nil, errors.New("must play one unseen card only")
			}
			// possible optimisation: could precalculate legal Unseen card moves
			cardIdx := msg.Decision[0]
			chosenCard := s.playerCards[s.currentPlayer.PlayerID].Unseen[cardIdx]

			s.completeMove(msg)

			legalMoves := getLegalMoves(s.pile, []deck.Card{chosenCard})

			if len(legalMoves) == 0 {
				s.pickUpPile()
				s.awaitingResponse = protocol.UnseenFailure
				return s.buildEndOfTurnMessages(protocol.UnseenFailure), nil
			}

			if s.playerHasFinished() {
				s.awaitingResponse = protocol.PlayerFinished
				return s.buildPlayerFinishedMessages(), nil
			}

			s.awaitingResponse = protocol.UnseenSuccess
			return s.buildEndOfTurnMessages(protocol.UnseenSuccess), nil
		}
	}

	return nil, errors.New("invalid game state")
}

// step 1 of 2 in a player playing their cards (Hand or Seen)
func (s *shed) attemptMove(currentPlayerCmd protocol.Cmd) ([]OutboundMessage, bool) {
	var cards []deck.Card
	switch currentPlayerCmd {

	case protocol.PlayHand:
		cards = s.playerCards[s.currentPlayer.PlayerID].Hand

	case protocol.PlaySeen:
		cards = s.playerCards[s.currentPlayer.PlayerID].Seen

	default:
		panic(fmt.Sprintf("unrecognised move protocol %s", currentPlayerCmd))
	}

	legalMoves := getLegalMoves(s.pile, cards)
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
		toPile      = []deck.Card{}
		newCards    *[]deck.Card
		newCardsSet = map[deck.Card]struct{}{}
	)

	switch msg.Command {
	case protocol.PlayHand:
		newCards = &s.playerCards[s.currentPlayer.PlayerID].Hand
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayer.PlayerID].Hand)

	case protocol.PlaySeen:
		newCards = &s.playerCards[s.currentPlayer.PlayerID].Seen
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayer.PlayerID].Seen)

	case protocol.PlayUnseen:
		newCards = &s.playerCards[s.currentPlayer.PlayerID].Unseen
		newCardsSet = cardSliceToSet(s.playerCards[s.currentPlayer.PlayerID].Unseen)
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
	s.playerCards[s.currentPlayer.PlayerID].Hand = append(s.playerCards[s.currentPlayer.PlayerID].Hand, fromDeck...)
}

func (s *shed) pickUpPile() {
	playerID := s.currentPlayer.PlayerID
	currentPlayerCards := s.playerCards[playerID]
	currentPlayerCards.Hand = append(currentPlayerCards.Hand, s.pile...)
	s.pile = []deck.Card{}
}

func (s *shed) playerHasFinished() bool {
	pc := s.playerCards[s.currentPlayer.PlayerID]
	return len(pc.Hand) == 0 &&
		len(pc.Seen) == 0 &&
		len(pc.Unseen) == 0
}

func (s *shed) gameIsOver() bool {
	return len(s.activePlayers) == 1
}

func (s *shed) turn() {
	s.currentTurnIdx = (s.currentTurnIdx + 1) % len(s.activePlayers)
	s.currentPlayer = s.activePlayers[s.currentTurnIdx]
}

func (s *shed) moveToFinishedPlayers() {
	if len(s.activePlayers) == 1 {
		s.finishedPlayers = append(s.finishedPlayers, s.activePlayers[0])
		s.activePlayers = []PlayerInfo{}
		return
	}

	s.activePlayers = append(s.activePlayers[:s.currentTurnIdx],
		s.activePlayers[(s.currentTurnIdx+1)%len(s.activePlayers):]...)

	s.finishedPlayers = append(s.finishedPlayers, s.currentPlayer)

	s.currentPlayer = s.activePlayers[s.currentTurnIdx]
}

func (s *shed) buildSkipTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.SkipTurn,
		CurrentTurn: s.currentPlayer,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
		Message:     fmt.Sprintf("%s skips a turn!", s.currentPlayer.Name),
	}
}

func (s *shed) buildSkipTurnMessages(currentPlayerCmd protocol.Cmd) []OutboundMessage {
	currentPlayerMsg := OutboundMessage{
		PlayerID:      s.currentPlayer.PlayerID,
		Command:       currentPlayerCmd, // always SkipTurn
		Hand:          s.playerCards[s.currentPlayer.PlayerID].Hand,
		Seen:          s.playerCards[s.currentPlayer.PlayerID].Seen,
		Pile:          s.pile,
		Opponents:     buildOpponents(s.currentPlayer.PlayerID, s.playerCards),
		ShouldRespond: true,
		Message:       "You skip a turn!",
	}

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, info := range s.playerInfo {
		if info.PlayerID != s.currentPlayer.PlayerID {
			toSend = append(toSend, s.buildSkipTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.Turn,
		CurrentTurn: s.currentPlayer,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
		Opponents:   buildOpponents(playerID, s.playerCards),
		Message:     fmt.Sprintf("It's %s's turn!", s.currentPlayer.Name),
	}
}

func (s *shed) buildTurnMessages(currentPlayerCmd protocol.Cmd, moves []int) []OutboundMessage {
	currentPlayerMsg := OutboundMessage{
		PlayerID:      s.currentPlayer.PlayerID,
		Command:       currentPlayerCmd,
		Hand:          s.playerCards[s.currentPlayer.PlayerID].Hand,
		Seen:          s.playerCards[s.currentPlayer.PlayerID].Seen,
		Pile:          s.pile,
		Moves:         moves,
		Opponents:     buildOpponents(s.currentPlayer.PlayerID, s.playerCards),
		ShouldRespond: true,
		Message:       "It's your turn!",
	}

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, info := range s.playerInfo {
		if info.PlayerID != s.currentPlayer.PlayerID {
			toSend = append(toSend, s.buildTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildEndOfTurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		Command:     protocol.EndOfTurn,
		CurrentTurn: s.currentPlayer,
		Hand:        s.playerCards[playerID].Hand,
		Seen:        s.playerCards[playerID].Seen,
		Pile:        s.pile,
	}
}

func (s *shed) buildEndOfTurnMessages(currentPlayerCommand protocol.Cmd) []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.playerInfo {
		msg := s.buildEndOfTurnMessage(info.PlayerID)
		if info.PlayerID == s.currentPlayer.PlayerID {
			msg.Command = currentPlayerCommand
			msg.ShouldRespond = true
		}
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildPlayerFinishedMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:        playerID,
		Command:         protocol.PlayerFinished,
		CurrentTurn:     s.currentPlayer,
		Hand:            s.playerCards[playerID].Hand,
		Seen:            s.playerCards[playerID].Seen,
		Pile:            s.pile,
		FinishedPlayers: s.finishedPlayers,
		Message:         fmt.Sprintf("%s has finished!", s.currentPlayer.Name),
	}
}

func (s *shed) buildPlayerFinishedMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.playerInfo {
		msg := s.buildPlayerFinishedMessage(info.PlayerID)
		if info.PlayerID == s.currentPlayer.PlayerID {
			msg.ShouldRespond = true
			msg.Message = "You've finished!"
		}
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildGameOverMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.playerInfo {
		toSend = append(toSend, OutboundMessage{
			PlayerID:        info.PlayerID,
			Command:         protocol.GameOver,
			FinishedPlayers: s.finishedPlayers,
			Pile:            s.pile,
			Message:         "Game over!",
		})
	}

	return toSend
}

func (s *shed) buildBurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:        playerID,
		Command:         protocol.Burn,
		CurrentTurn:     s.currentPlayer,
		Hand:            s.playerCards[playerID].Hand,
		Seen:            s.playerCards[playerID].Seen,
		Pile:            s.pile,
		FinishedPlayers: s.finishedPlayers,
		Message:         fmt.Sprintf("Burn for %s!", s.currentPlayer.Name),
	}
}

func (s *shed) buildBurnMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.playerInfo {
		msg := s.buildBurnMessage(info.PlayerID)
		if info.PlayerID == s.currentPlayer.PlayerID {
			msg.ShouldRespond = true
			msg.Message = "Burn!"
		}
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) getReorgCard(playerID string, choice int) deck.Card {
	oldHand := s.playerCards[playerID].Hand
	oldSeen := s.playerCards[playerID].Seen

	var card deck.Card
	if choice > 2 {
		card = oldSeen[choice-reorgSeenOffset]
	} else {
		card = oldHand[choice]
	}

	return card
}
