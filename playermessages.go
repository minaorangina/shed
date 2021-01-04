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
	Decision []int        `json:"decision,omitempty"` // not used in stage 0
}

// OutboundMessage is a message from GameEngine to Player
type OutboundMessage struct {
	PlayerID         string       `json:"player_id"`
	Name             string       `json:"name"` // pointless?
	Message          string       `json:"message"`
	Hand             []deck.Card  `json:"hand"`
	Seen             []deck.Card  `json:"seen"`
	Pile             []deck.Card  `json:"pile"`
	UnseenVisible    []deck.Card  `json:"unseen_visible",omitempty`
	CurrentTurn      string       `json:"current_turn",omitempty`
	Moves            []int        `json:"moves",omitempty`
	Opponents        []Opponent   `json:"opponents",omitempty`
	Command          protocol.Cmd `json:"command"`
	AwaitingResponse bool         `json:"awaiting_response"`
	Broadcast        bool
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
