package game

// Stage represents the game-specific stages.
type Stage int

const (
	preGame Stage = iota
	clearDeck
	clearCards // should I add a 'finished' stage?
)

// GamePlayState describes the current state of game play, i.e. game in progress or game over
type GamePlayState int

const (
	gameInProgress GamePlayState = iota
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
