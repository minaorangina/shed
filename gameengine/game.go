package gameengine

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
	for id, p := range ge.players {
		p.hand = confirmed[id].hand
		p.seen = confirmed[id].seen
	}

	return nil
}

func (ge *GameEngine) dealUnseenCards() {
	for _, p := range ge.players {
		p.unseen = ge.deck.Deal(3)
	}
}

func (ge *GameEngine) dealInitialCards() map[string]initialCards {
	cards := map[string]initialCards{}
	for _, p := range ge.players {
		dealtHand := ge.deck.Deal(3)
		dealtSeen := ge.deck.Deal(3)

		ic := initialCards{
			hand: dealtHand,
			seen: dealtSeen,
		}
		cards[p.ID] = ic
	}

	return cards
}

func (ge *GameEngine) confirmInitialCards(ic map[string]initialCards) (map[string]initialCards, error) {
	messages := OutboundMessages{}
	for _, p := range ge.players {
		o := buildOpponents(p.ID, ge.players)
		m := ge.buildReorgMessage(p, o, ic[p.ID], "Rearrange your hand")
		messages.Add(p.ID, m)
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
	player *Player,
	opponents []opponent,
	initialCards initialCards,
	message string,
) messageToPlayer {

	return messageToPlayer{
		PlayState: ge.playState,
		GameStage: ge.stage,
		PlayerID:  player.ID,
		Name:      player.Name,
		Message:   message,
		Hand:      initialCards.hand,
		Seen:      initialCards.seen,
		Opponents: opponents,
	}
}
