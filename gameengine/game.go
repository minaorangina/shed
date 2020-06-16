package gameengine

// Stage represents the main stages in the game
type Stage int

const (
	clearDeck Stage = iota
	clearCards
)

func (s Stage) String() string { // TODO: test
	if s == 0 {
		return "clearDeck"
	} else if s == 1 {
		return "clearCards"
	}
	return ""
}
