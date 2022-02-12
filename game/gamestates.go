package game

// Stage represents the main stages in the game
type Stage int

const (
	preGame Stage = iota
	clearDeck
	clearCards // should I add a 'finished' stage?
)

type GamePlayState int

const (
	gameNotStarted GamePlayState = iota
	gameStarted
	gameOver
)

type PlayerCardState int

const (
	playHand PlayerCardState = iota
	playSeen
	playUnseen
	empty
)

func validateStateMachine(s *shed) bool {
	// stages to validate
	// pre game
	// clear deck
	// - all players have three cards or more
	// - deck has at least one card
	// - no finished players
	// - expected commands: playhand, replenishhand, endofturn, skipturn, burn etc
	// clear cards
	// end game?
	return false
}
