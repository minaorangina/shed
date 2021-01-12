package protocol

// Cmd represents a command
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
