package gameengine

import (
	"math/rand"
	"time"
)

// playState represents the state of the current game
// idle -> no game play (pre game and post game)
// inProgress -> game in progress
// paused -> game is paused
type playState int

func (gps playState) String() string {
	if gps == 0 {
		return "idle"
	} else if gps == 1 {
		return "inProgress"
	} else if gps == 2 {
		return "paused"
	}
	return ""
}

const (
	idle playState = iota
	inProgress
	paused
)

type playerInfo struct {
	id   string
	name string
}

// GameEngine represents the engine of the game
type GameEngine struct {
	playState       playState
	externalPlayers map[string]ExternalPlayer
	game            *Game
}

// New constructs a new GameEngine
func New() GameEngine {
	return GameEngine{}
}

// Init initialises a new game
func (ge *GameEngine) Init(playerNames []string) error {
	if ge.playState != idle {
		return nil
	}

	info := make([]playerInfo, 0, len(playerNames))
	for _, name := range playerNames {
		info = append(info, playerInfo{name: name, id: NewID()})
	}

	shedGame, err := NewGame(ge, info)
	if err != nil {
		return err
	}

	// make external players
	external := map[string]ExternalPlayer{}
	for _, i := range info {
		external[i.id] = NewExternalPlayer(i.id, i.name)
	}

	ge.externalPlayers = external
	ge.game = shedGame
	ge.playState = inProgress

	ge.game.start()

	return nil
}

// PlayState returns the current gameplay status
func (ge *GameEngine) PlayState() string {
	return ge.playState.String()
}

func (ge *GameEngine) start() {
	ge.playState = inProgress
}

func (ge *GameEngine) messagePlayersAwaitReply(messages map[string]messageToPlayer) (map[string]messageFromPlayer, error) {
	cnl := make(chan messageFromPlayer)

	for _, msg := range messages {
		go ge.messagePlayer(cnl, ge.externalPlayers[msg.PlayerID], msg)
	}
	responses := map[string]messageFromPlayer{}
	for i := 0; i < len(messages); i++ {
		resp := <-cnl
		responses[resp.PlayerID] = resp
	}

	return responses, nil
}

func (ge *GameEngine) messagePlayer(cnl chan messageFromPlayer, externalPlayer ExternalPlayer, msg messageToPlayer) {
	rand.Seed(time.Now().UnixNano())
	timeout := rand.Intn(5)
	time.Sleep(time.Duration(100*timeout) * time.Millisecond)

	reply, err := externalPlayer.sendMessageAwaitReply(msg)
	if err != nil {
		// send default or something
	}
	_ = reply

	cnl <- messageFromPlayer{
		PlayerID:  msg.PlayerID,
		Command:   msg.Command,
		HandCards: msg.HandCards,
		SeenCards: msg.SeenCards,
	}
}
