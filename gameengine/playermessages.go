package gameengine

import "github.com/minaorangina/shed/deck"

type cmd int

const reorg cmd = iota

type OutboundMessage struct {
	PlayState playState   `json:"play_state"`
	GameStage Stage       `json:"game_stage"` // necessary?
	PlayerID  string      `json:"player_id"`
	Name      string      `json:"name"`
	Message   string      `json:"message"`
	Hand      []deck.Card `json:"hand"`
	Seen      []deck.Card `json:"seen"`
	Opponents []opponent  `json:"opponents"`
	Command   cmd         `json:"command"`
}

type InboundMessage struct {
	PlayerID string      `json:"player_id"`
	Command  cmd         `json:"command"`
	Hand     []deck.Card `json:"hand"`
	Seen     []deck.Card `json:"seen"`
}

// a response type
type initialCards struct {
	PlayerID string      `json:"player_id"`
	hand     []deck.Card `json:"hand"`
	seen     []deck.Card `json:"seen"`
}

// opponent is a representation of an opponent player
type opponent struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Seen []deck.Card `json:"opponent_seen"`
}
