package players

import (
	"os"
	"reflect"
	"testing"
)

func TestPlayer(t *testing.T) {
	name := "my name"
	p := NewPlayer("player-1", name, os.Stdin, os.Stdout)
	if p.Conn == nil {
		t.Fail()
	}
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

		ps = AddPlayer(&ps, APlayer(extraPlayerID, extraPlayerName))

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

		ps = AddPlayer(&ps, extraPlayer)
		ps = AddPlayer(&ps, extraPlayer)

		_, ok := ps.Find(extraPlayerID)

		if len(ps) != 3 {
			t.Errorf("Expected num Players to be 3, got %d", len(ps))
		}

		if !ok {
			t.Error("Could not add to Players")
		}
	})
}

// func TestNewPlayer(t *testing.T) {
// 	SomePlayers()
// }
