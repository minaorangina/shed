package gameengine

import (
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
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

// GameEngine represents the engine of the game
type GameEngine struct {
	playState  playState
	playerInfo []playerInfo
	game       *Game
}

// New constructs a new GameEngine
func New(playerNames []string) (GameEngine, error) {
	if len(playerNames) < 2 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: minimum of 2 players required (supplied %d)", len(playerNames))
	}
	if len(playerNames) > 4 {
		return GameEngine{}, fmt.Errorf("Could not construct GameEngine: maximum of 4 players allowed (supplied %d)", len(playerNames))
	}

	info := make([]playerInfo, 0, len(playerNames))
	for _, name := range playerNames {
		info = append(info, playerInfo{name: name, id: uuid.NewV4().String()})
	}

	return GameEngine{playerInfo: info}, nil
}

// Init initialises a new game
func (ge *GameEngine) Init() error {
	if ge.playState != idle {
		return nil
	}

	shedGame := NewGame(ge, ge.playerInfo)
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

func (ge *GameEngine) messagePlayersAwaitReply(messages []messageToPlayer) ([]reorganisedHand, error) {
	cnl := make(chan reorganisedHand)

	for _, msg := range messages {
		go ge.messagePlayer(cnl, "the-player", msg)
	}
	responses := []reorganisedHand{}
	for i := 0; i < len(messages); i++ {
		resp := <-cnl
		responses = append(responses, resp)
	}

	// send slice back
	return responses, nil
}

func (ge *GameEngine) messagePlayer(cnl chan reorganisedHand, player string, message messageToPlayer) {
	rand.Seed(time.Now().UnixNano())
	timeout := rand.Intn(5)
	time.Sleep(time.Duration(100*timeout) * time.Millisecond)

	// pass things down channel
	cnl <- reorganisedHand{
		PlayerID:  message.PlayerID,
		HandCards: message.HandCards,
		SeenCards: message.SeenCards,
	}
}
