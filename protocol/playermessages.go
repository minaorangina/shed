package protocol

import (
	"github.com/minaorangina/shed/deck"
)

type Player struct {
	PlayerID string `json:"playerID"`
	Name     string `json:"name"`
}

// InboundMessage is a message from Player to GameEngine
type InboundMessage struct {
	PlayerID string `json:"playerID"`
	Command  Cmd    `json:"command"`
	Decision []int  `json:"decision"`
}

// OutboundMessage is a message from GameEngine to Player
type OutboundMessage struct {
	PlayerID        string      `json:"playerID"`
	Command         Cmd         `json:"command"`
	Name            string      `json:"name"` // pointless?
	Message         string      `json:"message"`
	Hand            []deck.Card `json:"hand"`
	Seen            []deck.Card `json:"seen"`
	Unseen          []deck.Card `json:"unseen"`
	Pile            []deck.Card `json:"pile"`
	DeckCount       int         `json:"deckCount"`
	ShouldRespond   bool        `json:"shouldRespond"`
	Joiner          Player      `json:"joiner,omitempty"`
	CurrentTurn     Player      `json:"currentTurn,omitempty"`
	NextTurn        Player      `json:"nextTurn,omitempty"`
	Moves           []int       `json:"moves,omitempty"`
	Opponents       []Opponent  `json:"opponents,omitempty"`
	FinishedPlayers []Player    `json:"finishedPlayers,omitempty"`
	Error           string      `json:"error,omitempty"`
}

// Opponent is a representation of an opponent player
type Opponent struct {
	PlayerID string      `json:"playerID"`
	Name     string      `json:"name"`
	Seen     []deck.Card `json:"seen"`
}

type Cmd int

const (
	Null Cmd = iota
	NewJoiner
	Reorg
	Start
	HasStarted
	Error
	// combining game-specific and internal protocol messages.
	// will split later if necessary
	PlayHand      // when a player plays cards from their hand
	PlaySeen      // when a player plays cards from their seen cards
	PlayUnseen    // when a player plays cards from their unseen cards
	ReplenishHand // might disappear if EndOfTurn is better
	Turn
	EndOfTurn
	SkipTurn
	Burn
	UnseenSuccess
	UnseenFailure
	PlayerFinished
	GameOver
)

var CmdNames = map[Cmd]string{
	Null:           "Null",
	NewJoiner:      "NewJoiner",
	Reorg:          "Reorg",
	Start:          "Start",
	HasStarted:     "HasStarted",
	Error:          "Error",
	PlayHand:       "PlayHand",
	PlaySeen:       "PlaySeen",
	PlayUnseen:     "PlayUnseen",
	ReplenishHand:  "ReplenishHand",
	Turn:           "Turn",
	EndOfTurn:      "EndOfTurn",
	SkipTurn:       "SkipTurn",
	Burn:           "Burn",
	UnseenSuccess:  "UnseenSuccess",
	UnseenFailure:  "UnseenFailure",
	PlayerFinished: "PlayerFinished",
	GameOver:       "GameOver",
}

var NameToCmd = map[string]Cmd{
	"Null":           Null,
	"NewJoiner":      NewJoiner,
	"Reorg":          Reorg,
	"Start":          Start,
	"HasStarted":     HasStarted,
	"Error":          Error,
	"PlayHand":       PlayHand,
	"PlaySeen":       PlaySeen,
	"PlayUnseen":     PlayUnseen,
	"ReplenishHand":  ReplenishHand,
	"Turn":           Turn,
	"EndOfTurn":      EndOfTurn,
	"SkipTurn":       SkipTurn,
	"Burn":           Burn,
	"UnseenSuccess":  UnseenSuccess,
	"UnseenFailure":  UnseenFailure,
	"PlayerFinished": PlayerFinished,
	"GameOver":       GameOver,
}

func (c Cmd) String() string {
	return CmdNames[c]
}
