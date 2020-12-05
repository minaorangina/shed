package shed

import "github.com/minaorangina/shed/protocol"

/**
 * Here be stuff I haven't clean up yet.
 */

type CLIPlayer struct {
	PlayerCards
	id   string
	name string
	Conn *conn
}

func (p CLIPlayer) ID() string {
	return p.id
}

func (p CLIPlayer) Name() string {
	return p.name
}

// Cards returns all of a player's cards
func (p CLIPlayer) Cards() *PlayerCards {
	return &PlayerCards{
		Hand:   p.Hand,
		Seen:   p.Seen,
		Unseen: p.Unseen,
	}
}

func (p CLIPlayer) Send(msg OutboundMessage) error {
	// check command
	switch msg.Command {
	case protocol.Reorg:
		inbound := p.handleReorg(msg)
		_ = inbound

		// convert to appropriate format

		// call Send on the connection

		return nil
	}
	return nil
}

func (p CLIPlayer) Receive(msg []byte) {
	// convert to InboundMessage

	// put on the game engine chan
}

func (p CLIPlayer) handleReorg(msg OutboundMessage) InboundMessage {
	response := InboundMessage{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Hand:     msg.Hand,
		Seen:     msg.Seen,
	}

	playerCards := PlayerCards{
		Seen: msg.Seen,
		Hand: msg.Hand,
	}
	SendText(p.Conn.Out, "%s, here are your cards:\n\n", msg.Name)
	SendText(p.Conn.Out, buildCardDisplayText(playerCards))

	// could probably make this one step. user submits or skips - no need to
	if shouldReorganise := offerCardSwitch(p.Conn, offerTimeout); shouldReorganise {
		response = reorganiseCards(p.Conn, msg)
	}

	return response
}
