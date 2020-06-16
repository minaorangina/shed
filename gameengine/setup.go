package gameengine

import "github.com/minaorangina/shed/players"

// HandleInitialCards is the setup function for Shed
func HandleInitialCards(ge GameEngine) error {
	// shuffle
	deck := ge.Deck()
	deck.Shuffle()

	// deal
	dealUnseenCards(ge)
	initial := dealInitialCards(ge)

	// confirm with player
	confirmed, err := confirmInitialCards(ge, initial)
	if err != nil {
		return err
	}

	// assign cards
	for _, p := range ge.Players() {
		p.Hand = confirmed[p.ID].Hand
		p.Seen = confirmed[p.ID].Seen
	}

	return nil
}

func dealUnseenCards(ge GameEngine) {
	deck := ge.Deck()
	for _, p := range ge.Players() {
		p.Unseen = deck.Deal(3)
	}
}

func dealInitialCards(ge GameEngine) map[string]players.InitialCards {
	cards := map[string]players.InitialCards{}
	deck := ge.Deck()
	for _, p := range ge.Players() {
		dealtHand := deck.Deal(3)
		dealtSeen := deck.Deal(3)

		ic := players.InitialCards{
			Hand: dealtHand,
			Seen: dealtSeen,
		}
		cards[p.ID] = ic
	}

	return cards
}

func confirmInitialCards(ge GameEngine, ic map[string]players.InitialCards) (map[string]players.InitialCards, error) {
	messages := []players.OutboundMessage{}
	for _, p := range ge.Players() {
		o := buildOpponents(p.ID, ge.Players())
		m := buildReorgMessage(p, o, ic[p.ID], "Rearrange your hand")
		messages = append(messages, m)
	}

	// this will block
	reply, err := ge.MessagePlayers(messages)
	if err != nil {
		return nil, err
	}

	return messagesToInitialCards(reply), nil
}

// to test (easier when state hydration exists)
func buildReorgMessage(
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
