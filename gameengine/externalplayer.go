package gameengine

import "os"

type conn struct {
	In  *os.File
	Out *os.File
}

// ExternalPlayer represents a player in the outside world
// An ExternalPlayer maps onto a Player in the game
type ExternalPlayer struct {
	ID   string
	Name string
	Conn *conn // tcp or command line
}

// NewExternalPlayer constructs an ExternalPlayer
func NewExternalPlayer(id, name string) ExternalPlayer {
	conn := &conn{os.Stdin, os.Stdout}
	return ExternalPlayer{
		ID:   id,
		Name: name,
		Conn: conn,
	}
}
