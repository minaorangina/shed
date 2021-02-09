package shed

import (
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
)

// InboundMessage is a message from Player to GameEngine
type InboundMessage struct {
	PlayerID string       `json:"playerID"`
	Command  protocol.Cmd `json:"command"`
	Decision []int        `json:"decision"`
}

// OutboundMessage is a message from GameEngine to Player
type OutboundMessage struct {
	PlayerID        string       `json:"playerID"`
	Command         protocol.Cmd `json:"command"`
	Name            string       `json:"name"` // pointless?
	Joiner          PlayerInfo   `json:"joiner,omitempty"`
	Message         string       `json:"message"`
	Hand            []deck.Card  `json:"hand"`
	Seen            []deck.Card  `json:"seen"`
	Pile            []deck.Card  `json:"pile"`
	CurrentTurn     PlayerInfo   `json:"currentTurn,omitempty"`
	Moves           []int        `json:"moves,omitempty"`
	Opponents       []Opponent   `json:"opponents,omitempty"`
	FinishedPlayers []PlayerInfo `json:"finishedPlayers,omitempty"`
	ShouldRespond   bool         `json:"shouldRespond"`
	Error           string       `json:"error,omitempty"`
}

// InitialCards represent the default cards dealt to a Player
type InitialCards struct {
	PlayerID string      `json:"playerID"`
	Hand     []deck.Card `json:"hand"`
	Seen     []deck.Card `json:"seen"`
}

// Opponent is a representation of an opponent player
type Opponent struct {
	PlayerID string      `json:"playerID"`
	Name     string      `json:"name"`
	Seen     []deck.Card `json:"seen"`
}
