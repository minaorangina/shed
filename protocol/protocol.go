package protocol

// Cmd represents a command
type Cmd int

const (
	NewJoiner Cmd = iota
	Reorg
	Start
	HasStarted
	// combining game-specific and internal protocol messages.
	// will split later if necessary
	PlayHand      // when a player plays cards from their hand
	PlaySeen      // when a player plays cards from their seen cards
	PlayUnseen    // when a player plays cards from their unseen cards
	ReplenishHand // might disappear if EndOfTurn is better
	Turn
	EndOfTurn
	SkipTurn
	UnseenSuccess
	UnseenFailure
	PlayerFinished
	GameOver
)

var cmdNames = []string{
	"NewJoiner",
	"Reorg",
	"Start",
	"HasStarted",
	"PlayHand",
	"PlaySeen",
	"PlayUnseen",
	"ReplenishHand",
	"Turn",
	"EndOfTurn",
	"SkipTurn",
	"UnseenSuccess",
	"UnseenFailure",
	"PlayerFinished",
	"GameOver",
}

func (c Cmd) String() string {
	return cmdNames[c]
}
