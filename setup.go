package shed

import (
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
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

	// // confirm with player
	// err := confirmInitialCards(ps, initialCards, ge.MessagePlayers)
	// if err != nil {
	// 	return err
	// }

	// prompt player
	messages := []OutboundMessage{}
	for _, p := range ps {
		playerID := p.ID()
		o := buildOpponents(playerID, ps)
		m := buildReorgMessage(p, o, initialCards[playerID])
		messages = append(messages, m)
	}

	err := ge.MessagePlayers(messages)
	if err != nil {
		return err
	}

	// // block to get initial cards here or something
	// confirmed := map[string]PlayerCards{}

	// // assign cards
	// for _, p := range ps {
	// 	cards := p.Cards()
	// 	cards.Hand = confirmed[p.ID()].Hand
	// 	cards.Seen = confirmed[p.ID()].Seen
	// }

	return nil
}

func dealUnseenCards(deck deck.Deck, ps Players) {
	for _, p := range ps {
		cards := p.Cards()
		cards.Unseen = deck.Deal(3)
	}
}

func dealInitialCards(deck deck.Deck, ps Players) map[string]InitialCards {
	cards := map[string]InitialCards{}
	for _, p := range ps {
		dealtHand := deck.Deal(3)
		dealtSeen := deck.Deal(3)

		ic := InitialCards{
			Hand: dealtHand,
			Seen: dealtSeen,
		}
		cards[p.ID()] = ic
	}

	return cards
}

func confirmInitialCards(
	ps Players,
	ic map[string]InitialCards,
	messageFn func([]OutboundMessage) error,
) error {
	messages := []OutboundMessage{}
	for _, p := range ps {
		playerID := p.ID()
		o := buildOpponents(playerID, ps)
		m := buildReorgMessage(p, o, ic[playerID])
		messages = append(messages, m)
	}

	// this will block
	return messageFn(messages)
}

// to test (easier when state hydration exists)
func buildReorgMessage(
	player Player,
	opponents []Opponent,
	initialCards InitialCards,
) OutboundMessage {

	return OutboundMessage{
		PlayerID:  player.ID(),
		Name:      player.Name(),
		Hand:      initialCards.Hand,
		Seen:      initialCards.Seen,
		Opponents: opponents,
		Command:   protocol.Reorg,
	}
}
