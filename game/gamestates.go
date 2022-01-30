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
