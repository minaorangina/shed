package players

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type TestConn struct {
	In  io.Reader
	Out io.Writer
}

type TestPlayer struct {
	cards    *PlayerCards
	id       string
	name     string
	conn     *TestConn
	received []byte
}

func NewTestPlayer(id, name string, in io.Reader, out io.Writer) *TestPlayer {
	return &TestPlayer{
		id:    id,
		name:  name,
		conn:  &TestConn{in, out},
		cards: &PlayerCards{},
	}
}

func (tp *TestPlayer) ID() string {
	return tp.id
}

func (tp *TestPlayer) Name() string {
	return tp.name
}

func (tp *TestPlayer) Cards() *PlayerCards {
	return tp.cards
}

func (tp *TestPlayer) Send(msg OutboundMessage) error {
	fmt.Fprintf(tp.conn.Out, ("hello"))
	return nil
}

func (tp *TestPlayer) Receive(data []byte) {
	tp.received = data
}

func APlayer(id, name string) Player {
	// does bytes.Buffer need to change to NewTestBuffer?
	return NewTestPlayer(id, name, &bytes.Buffer{}, ioutil.Discard)
}

func SomePlayers() Players {
	player1 := NewTestPlayer(NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := NewTestPlayer(NewID(), "Sally", os.Stdin, os.Stdout)
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
