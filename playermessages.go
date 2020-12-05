package shed

import (
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
)

// InboundMessage is a message from Player to GameEngine
type InboundMessage struct {
	PlayerID string       `json:"player_id"`
	Command  protocol.Cmd `json:"command"`
	Hand     []deck.Card  `json:"hand",omitempty`
	Seen     []deck.Card  `json:"seen",omitempty`
}

// OutboundMessage is a message from GameEngine to Player
type OutboundMessage struct {
	PlayerID  string       `json:"player_id"`
	Name      string       `json:"name"` // pointless?
	Message   string       `json:"message"`
	Hand      []deck.Card  `json:"hand"`
	Seen      []deck.Card  `json:"seen"`
	Opponents []Opponent   `json:"opponents"`
	Command   protocol.Cmd `json:"command"`
	Broadcast bool
}

// InitialCards represent the default cards dealt to a Player
type InitialCards struct {
	PlayerID string      `json:"player_id"`
	Hand     []deck.Card `json:"hand"`
	Seen     []deck.Card `json:"seen"`
}

// Opponent is a representation of an opponent player
type Opponent struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Seen []deck.Card `json:"opponent_seen"`
}
