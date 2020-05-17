package gameengine

import "github.com/minaorangina/shed/deck"

type cmd int

const reorg cmd = iota

type messageToPlayer struct {
	PlayState playState   `json:"play_state"`
	GameStage Stage       `json:"game_stage"` // necessary?
	PlayerID  string      `json:"player_id"`
	Name      string      `json:"name"`
	Message   string      `json:"message"`
	HandCards []deck.Card `json:"hand_cards"`
	SeenCards []deck.Card `json:"seen_cards"`
	Opponents []opponent  `json:"opponents"`
	Command   cmd         `json:"command"`
}

type messageFromPlayer struct {
	PlayerID  string      `json:"player_id"`
	Command   cmd         `json:"command"`
	HandCards []deck.Card `json:"hand_cards"`
	SeenCards []deck.Card `json:"seen_cards"`
}

// a response type
type reorganisedHand struct {
	PlayerID  string      `json:"player_id"`
	HandCards []deck.Card `json:"hand_cards"`
	SeenCards []deck.Card `json:"seen_cards"`
}

// opponent is a representation of an opponent player
type opponent struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	SeenCards []deck.Card `json:"opponent_seen_cards"`
}
