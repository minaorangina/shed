package shed

import (
	"reflect"
	"testing"
)

type spyWSConn struct {
	calls [][]byte
}

func (c *spyWSConn) Send(data []byte) error {
	c.calls = append(c.calls, data)
	return nil
}

func (c *spyWSConn) Receive(data []byte) {}

func TestWSPlayer(t *testing.T) {
	// test that calling Send() calls conn.Send()
	// t.Run("Send passes message on to connection", func(t *testing.T) {
	// 	spy := &spyWSConn{} // mock with a mocking library?
	// 	player := &WSPlayer{id: "an-id", name: "a name", conn: spy}

	// 	err := player.Send(OutboundMessage{Message: "Amber"})
	// 	utils.AssertNoError(t, err)
	// 	utils.AssertTrue(t, len(spy.calls) > 0)
	// })
}

func TestWSConn(t *testing.T) {

}

func TestPlayers(t *testing.T) {
	t.Run("Can find a player", func(t *testing.T) {
		playerID := "a-player-id"

		player1 := APlayer(playerID, "Orville")

		ps := NewPlayers(player1)

		got, ok := ps.Find(playerID)

		if !ok {
			t.Error("Failed to retrieve a player")
		}
		if !reflect.DeepEqual(got, player1) {
			t.Error("Failed to find the correct player")
		}
	})

	t.Run("Can add more players", func(t *testing.T) {
		ps := NewPlayers(SomePlayers()...)

		extraPlayerID := "another-player"
		extraPlayerName := "Human"

		ps = AddPlayer(ps, APlayer(extraPlayerID, extraPlayerName))

		_, ok := ps.Find(extraPlayerID)

		if !ok {
			t.Error("Could not add to Players")
		}
	})

	t.Run("Adding an existing player is a no-op", func(t *testing.T) {
		ps := NewPlayers(SomePlayers()...)

		extraPlayerID := "another-player"
		extraPlayerName := "Human"
		extraPlayer := APlayer(extraPlayerID, extraPlayerName)

		ps = AddPlayer(ps, extraPlayer)
		ps = AddPlayer(ps, extraPlayer)

		_, ok := ps.Find(extraPlayerID)

		if len(ps) != 3 {
			t.Errorf("Expected num Players to be 3, got %d", len(ps))
		}

		if !ok {
			t.Error("Could not add to Players")
		}
	})
}
