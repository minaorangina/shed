package gameengine

// Stage represents the main stages in the game
type Stage int

const (
	cardOrganisation Stage = iota
	clearDeck
	clearCards
)

func (s Stage) String() string {
	if s == 0 {
		return "cardOrganisation"
	} else if s == 1 {
		return "clearDeck"
	} else if s == 2 {
		return "clearCards"
	}
	return ""
}
