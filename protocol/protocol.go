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
	PlayHand // when a player plays cards from their hand
	ReplenishHand
	Turn
)

var cmdNames = []string{
	"NewJoiner",
	"Reorg",
	"Start",
	"HasStarted",
	"NoLegalMoves",
	"PlayHand",
	"ReplenishHand",
	"Turn",
}

func (c Cmd) String() string {
	return cmdNames[c]
}
