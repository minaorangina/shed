package gameengine

import (
	"errors"
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
)

type spySetup struct {
	called bool
}

func (s *spySetup) setup(ge GameEngine) error {
	s.called = true
	return nil
}

func TestGameEngineSetupFn(t *testing.T) {
	t.Run("starts correctly", func(t *testing.T) {
		spy := spySetup{}
		engine, err := New(somePlayers(), spy.setup)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		utils.AssertNoError(t, err)

		if spy.called != true {
			t.Errorf("Expected spy setup fn to be called")
		}
	})

	t.Run("does not error if no setup fn defined", func(t *testing.T) {
		engine, err := New(somePlayers(), nil)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		utils.AssertNoError(t, err)
	})

	t.Run("propagates setup fn error", func(t *testing.T) {
		erroringSetupFn := func(ge GameEngine) error {
			return errors.New("Whoops")
		}
		engine, err := New(somePlayers(), erroringSetupFn)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		if err == nil {
			t.Fatalf("Expected an error, but there was none")
		}
	})
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// Game Engine receives from messages to send to players
	// and returns response
	t.Skip("do not run TestGameEngineMsgFromGame")
	ge, _ := gameEngineWithPlayers()
	ge.Start() // mock required

	messages := []players.OutboundMessage{}
	expected := []players.InboundMessage{}
	initialCards := players.InitialCards{}
	for _, p := range ge.Players() {
		o := buildOpponents(p.ID, ge.Players())
		m := buildReorgMessage(p, o, initialCards, "Rearrange your initial cards")
		messages = append(messages, m)

		expected = append(expected, players.InboundMessage{
			PlayerID: p.ID,
			Hand:     p.Hand,
			Seen:     p.Seen,
		})
	}

	got, err := ge.MessagePlayers(messages)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(got, want) { // TODO: fix slice ordering
		utils.FailureMessage(t, got, want)
	}
}
