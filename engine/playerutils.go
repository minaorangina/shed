package engine

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/minaorangina/shed/game"
	"github.com/minaorangina/shed/protocol"
)

type FakeConn struct {
	In   io.Reader
	Out  io.Writer
	sent [][]byte
}

func (c *FakeConn) Send(data []byte) error {
	c.sent = append(c.sent, data)
	return nil
}

func (c *FakeConn) Receive(data []byte) {
	// do something
}

type TestPlayer struct {
	cards    *game.PlayerCards
	id       string
	name     string
	conn     *FakeConn
	received []byte
}

func NewTestPlayer(id, name string, in io.Reader, out io.Writer) *TestPlayer {
	return &TestPlayer{
		id:    id,
		name:  name,
		conn:  &FakeConn{in, out, [][]byte{}},
		cards: &game.PlayerCards{},
	}
}

func (tp *TestPlayer) Info() protocol.PlayerInfo {
	return protocol.PlayerInfo{
		PlayerID: tp.id,
		Name:     tp.name,
	}
}

func (tp *TestPlayer) ID() string {
	return tp.id
}

func (tp *TestPlayer) Name() string {
	return tp.name
}

func (tp *TestPlayer) Cards() *game.PlayerCards {
	return tp.cards
}

func (tp *TestPlayer) Send(msg protocol.OutboundMessage) error {
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

// TestBuffer is used in tests for io
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

func charsUnique(s string) bool {
	seen := map[string]bool{}
	for _, c := range s {
		if _, ok := seen[string(c)]; ok {
			return false
		}
		seen[string(c)] = true
	}
	return true
}

func charsInRange(chars string, lower, upper int) bool {
	for _, char := range chars {
		if int(char) < lower || int(char) > upper {
			return false
		}
	}
	return true
}
