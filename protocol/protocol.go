package protocol

// Cmd represents a command
type Cmd int

const (
	NewJoiner Cmd = iota
	Reorg
)
