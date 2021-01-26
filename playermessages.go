package shed

import (
	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/protocol"
)

// InboundMessage is a message from Player to GameEngine
type InboundMessage struct {
	PlayerID string       `json:"player_id"`
	Command  protocol.Cmd `json:"command"`
	Hand     []deck.Card  `json:"hand,omitempty"`
	Seen     []deck.Card  `json:"seen,omitempty"`
	Decision []int        `json:"decision,omitempty"` // not used in stage 0
}

// func (im *InboundMessage) MarshalJSON() ([]byte, error) {
// 	type Alias InboundMessage
// 	return json.Marshal(&struct {
// 		Command string `json:"command"`
// 		*Alias
// 	}{
// 		Command: im.Command.String(),
// 		Alias:   (*Alias)(im),
// 	})
// }

// func (im *InboundMessage) UnmarshalJSON(data []byte) error {
// 	type Alias InboundMessage
// 	aux := &struct {
// 		Command string `json:"command"`
// 		*Alias
// 	}{
// 		Alias: (*Alias)(im),
// 	}
// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}
// 	v, ok := protocol.NameToCmd[aux.Command]
// 	if !ok {
// 		return fmt.Errorf("unknown command %s", aux.Command)
// 	}

// 	im.Command = v
// 	return nil
// }

// OutboundMessage is a message from GameEngine to Player
type OutboundMessage struct {
	PlayerID        string       `json:"player_id"`
	Command         protocol.Cmd `json:"command"`
	Name            string       `json:"name"` // pointless?
	Joiner          string       `json:"joiner,omitempty"`
	Message         string       `json:"message"`
	Hand            []deck.Card  `json:"hand"`
	Seen            []deck.Card  `json:"seen"`
	Pile            []deck.Card  `json:"pile"`
	CurrentTurn     string       `json:"current_turn,omitempty"`
	Moves           []int        `json:"moves,omitempty"`
	Opponents       []Opponent   `json:"opponents,omitempty"`
	FinishedPlayers []string     `json:"finished_players,omitempty"`
	ShouldRespond   bool         `json:"should_respond"`
	Error           string       `json:"error,omitempty"`
	DefaultMessage  InboundMessage
}

// InitialCards represent the default cards dealt to a Player
type InitialCards struct {
	PlayerID string      `json:"player_id"`
	Hand     []deck.Card `json:"hand"`
	Seen     []deck.Card `json:"seen"`
}

// Opponent is a representation of an opponent player
type Opponent struct {
	PlayerID string      `json:"player_id"`
	Name     string      `json:"name"`
	Seen     []deck.Card `json:"opponent_seen"`
}
