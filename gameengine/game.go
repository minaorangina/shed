package gameengine

import "github.com/minaorangina/shed/players"

// Stage represents the main stages in the game
type Stage int

const (
	cardOrganisation Stage = iota
	clearDeck
	clearCards
)

func (s Stage) String() string {
	if s == 0 {
		return "cardOrganisation"
	} else if s == 1 {
		return "clearDeck"
	} else if s == 2 {
		return "clearCards"
	}
	return ""
}

func (ge *GameEngine) handleInitialCards() error {
	// shuffle
	ge.deck.Shuffle()

	// deal
	ge.dealUnseenCards()
	initial := ge.dealInitialCards()

	// confirm with player
	confirmed, err := ge.confirmInitialCards(initial)
	if err != nil {
		return err
	}

	// assign cards
	for _, p := range ge.players {
		p.Hand = confirmed[p.ID].Hand
		p.Seen = confirmed[p.ID].Seen
	}

	return nil
}

func (ge *GameEngine) dealUnseenCards() {
	for _, p := range ge.players {
		p.Unseen = ge.deck.Deal(3)
	}
}

func (ge *GameEngine) dealInitialCards() map[string]players.InitialCards {
	cards := map[string]players.InitialCards{}
	for _, p := range ge.players {
		dealtHand := ge.deck.Deal(3)
		dealtSeen := ge.deck.Deal(3)

		ic := players.InitialCards{
			Hand: dealtHand,
			Seen: dealtSeen,
		}
		cards[p.ID] = ic
	}

	return cards
}

func (ge *GameEngine) confirmInitialCards(ic map[string]players.InitialCards) (map[string]players.InitialCards, error) {
	messages := []players.OutboundMessage{}
	for _, p := range ge.players {
		o := buildOpponents(p.ID, ge.players)
		m := ge.buildReorgMessage(p, o, ic[p.ID], "Rearrange your hand")
		messages = append(messages, m)
	}

	// this will block
	reply, err := ge.messagePlayersAwaitReply(messages)
	if err != nil {
		return nil, err
	}

	return messagesToInitialCards(reply), nil
}

// to test (easier when state hydration exists)
func (ge *GameEngine) buildReorgMessage(
	player *players.Player,
	opponents []players.Opponent,
	initialCards players.InitialCards,
	message string,
) players.OutboundMessage {

	return players.OutboundMessage{
		PlayerID:  player.ID,
		Name:      player.Name,
		Message:   message,
		Hand:      initialCards.Hand,
		Seen:      initialCards.Seen,
		Opponents: opponents,
	}
}
