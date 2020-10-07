package players

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
)

func APlayer(id, name string) *Player {
	// does bytes.Buffer need to change to NewTestBuffer?
	return NewPlayer(id, name, &bytes.Buffer{}, ioutil.Discard)
}

func SomePlayers() Players {
	player1 := NewPlayer(NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := NewPlayer(NewID(), "Sally", os.Stdin, os.Stdout)
	players := NewPlayers(player1, player2)
	return players
}

type TestBuffer struct {
	buf bytes.Buffer
	m   sync.Mutex
}

func NewTestBuffer() *TestBuffer {
	return &TestBuffer{}
}

func (tb *TestBuffer) Read(p []byte) (int, error) {
	tb.m.Lock()
	defer tb.m.Unlock()
	return tb.buf.Read(p)
}

func (tb *TestBuffer) Write(p []byte) (int, error) {
	tb.m.Lock()
	defer tb.m.Unlock()
	return tb.buf.Write(p)
}

func (tb *TestBuffer) String() string {
	tb.m.Lock()
	defer tb.m.Unlock()
	return tb.buf.String()
}
