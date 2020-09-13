package shed

import (
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/players"
)

// HandleInitialCards is the setup function for Shed
func HandleInitialCards(ge GameEngine) error {
	deck := ge.Deck()
	ps := ge.Players()

	// shuffle
	deck.Shuffle()

	// deal
	dealUnseenCards(deck, ps)
	initialCards := dealInitialCards(deck, ps)

	// confirm with player
	confirmed, err := confirmInitialCards(ps, initialCards, ge.MessagePlayers)
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

func dealUnseenCards(deck deck.Deck, ps players.Players) {
	for _, p := range ps {
		p.Unseen = deck.Deal(3)
	}
}

func dealInitialCards(deck deck.Deck, ps players.Players) map[string]players.InitialCards {
	cards := map[string]players.InitialCards{}
	for _, p := range ps {
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

func confirmInitialCards(
	ps players.Players,
	ic map[string]players.InitialCards,
	messageFn func([]players.OutboundMessage) ([]players.InboundMessage, error),
) (map[string]players.InitialCards, error) {
	messages := []players.OutboundMessage{}
	for _, p := range ps {
		o := buildOpponents(p.ID, ps)
		m := buildReorgMessage(p, o, ic[p.ID], "Rearrange your hand")
		messages = append(messages, m)
	}

	// this will block
	reply, err := messageFn(messages)
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
