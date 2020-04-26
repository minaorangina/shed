package externalplayer

// ExternalPlayer represents a player in the outside world
// An ExternalPlayer maps onto a Player in the game
type ExternalPlayer struct {
	id   int
	name string
	conn interface{} // tcp or command line
}
