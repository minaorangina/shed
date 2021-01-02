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
	NoLegalMoves
	PlayHand      // when a player plays cards from their hand
	PlaySeen      // when a player plays cards from their seen cards
	PlayUnseen    // when a player plays cards from their seen cards
	ReplenishHand // might disappear if EndOfTurn is better
	Turn
	EndOfTurn
)

var cmdNames = []string{
	"NewJoiner",
	"Reorg",
	"Start",
	"HasStarted",
	"NoLegalMoves",
	"PlayHand",
	"PlaySeen",
	"PlayUnseen",
	"ReplenishHand",
	"Turn",
	"EndOfTurn",
}

func (c Cmd) String() string {
	return cmdNames[c]
}
