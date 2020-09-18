package players

import (
	"bytes"
	"io/ioutil"
	"os"
)

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

func APlayer(id, name string) *Player {
	return NewPlayer(id, name, &bytes.Buffer{}, ioutil.Discard)
}

func SomePlayers() Players {
	player1 := NewPlayer(NewID(), "Harry", os.Stdin, os.Stdout)
	player2 := NewPlayer(NewID(), "Sally", os.Stdin, os.Stdout)
	players := NewPlayers(player1, player2)
	return players
}
