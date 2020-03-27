package player

// Player represents a player in the game
type Player struct {
	name string
}

// New constructs a new player
func New(name string) (Player, error) {
	return Player{name}, nil
}
