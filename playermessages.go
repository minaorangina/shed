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
	Message         string       `json:"message"`
	Hand            []deck.Card  `json:"hand"`
	Seen            []deck.Card  `json:"seen"`
	Pile            []deck.Card  `json:"pile"`
	DeckCount       int          `json:"deckCount"`
	ShouldRespond   bool         `json:"shouldRespond"`
	Joiner          PlayerInfo   `json:"joiner,omitempty"`
	CurrentTurn     PlayerInfo   `json:"currentTurn,omitempty"`
	Moves           []int        `json:"moves,omitempty"`
	Opponents       []Opponent   `json:"opponents,omitempty"`
	FinishedPlayers []PlayerInfo `json:"finishedPlayers,omitempty"`
	Error           string       `json:"error,omitempty"`
}

// Opponent is a representation of an opponent player
type Opponent struct {
	PlayerID string      `json:"playerID"`
	Name     string      `json:"name"`
	Seen     []deck.Card `json:"seen"`
}
