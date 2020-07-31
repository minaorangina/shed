package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertNewGameResponse(t *testing.T, body *bytes.Buffer, want string) {
	t.Helper()
	bodyBytes, _ := ioutil.ReadAll(body)

	var got NewGameRes
	err := json.Unmarshal(bodyBytes, &got)
	if err != nil {
		t.Fatalf("Could not unmarshal json: %s", err.Error())
	}
	if got.Name != want {
		t.Errorf("Got %s, want %s", got.Name, want)
	}
	if len(got.GameID) == 0 {
		t.Error("Expected a game id")
	}
	if len(got.PlayerID) == 0 {
		t.Error("Expected a player id")
	}
}
