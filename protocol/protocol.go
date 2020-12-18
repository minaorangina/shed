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
	PlayHand
	Replenish
)
