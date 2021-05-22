package game

// Stage represents the main stages in the game
type Stage int

const (
	preGame Stage = iota
	clearDeck
	clearCards
)

type GamePlayState int

const (
	notStarted GamePlayState = iota
	started
	finished
)
