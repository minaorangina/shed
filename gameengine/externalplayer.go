package gameengine

import (
	"fmt"
	"os"
)

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
func NewExternalPlayer(id, name string, in, out *os.File) ExternalPlayer {
	conn := &conn{in, out}
	return ExternalPlayer{
		ID:   id,
		Name: name,
		Conn: conn,
	}
}

func (ep ExternalPlayer) sendMessageAwaitReply(msg messageToPlayer) (messageFromPlayer, error) {
	var response messageFromPlayer

	switch msg.Command {
	case reorg:
		resp, err := ep.handleReorg(msg)
		if err != nil {
			return messageFromPlayer{}, err
		}
		response = resp
	}
	fmt.Println("GAMEENGINE", response)
	return response, nil
}
