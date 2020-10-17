package shed

import (
	"errors"
	"fmt"
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

func TestGameEngineAddPlayer(t *testing.T) {
	t.Run("can add more players", func(t *testing.T) {
		playerID := "some-player-id"
		name := "Héloise"
		ge := gameEngineWithPlayers()

		err := ge.AddPlayer(players.APlayer(playerID, name))
		utils.AssertNoError(t, err)

		ps := ge.Players()
		_, ok := ps.Find(playerID)
		utils.AssertTrue(t, ok)
	})
}

func TestGameEngineInit(t *testing.T) {
	t.Run("constructs with correct number of cards", func(t *testing.T) {
		ge := gameEngineWithPlayers()
		if len(ge.Deck()) != 52 {
			t.Errorf("\nExpected: %+v\nActual: %+v\n", 52, len(ge.Deck()))
		}
	})
	t.Run("has an ID", func(t *testing.T) {
		id := "thisistheid"
		engine, err := New(id, players.SomePlayers(), nil)
		utils.AssertNoError(t, err)

		utils.AssertEqual(t, engine.ID(), id)
	})
}
func TestGameEngineSetupFn(t *testing.T) {
	t.Run("sets up correctly", func(t *testing.T) {
		spy := spySetup{}
		engine, err := New("", players.SomePlayers(), spy.setup)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		utils.AssertNoError(t, err)

		if spy.called != true {
			t.Errorf("Expected spy setup fn to be called")
		}
	})

	t.Run("requires legal number of players", func(t *testing.T) {
		type gameTest struct {
			testName string
			input    players.Players
			want     error
		}
		testsShouldError := []gameTest{
			{
				"too few players",
				namesToPlayers([]string{"Grace"}),
				ErrTooFewPlayers,
			},
			{
				"too many players",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
				ErrTooManyPlayers,
			},
		}

		for _, et := range testsShouldError {
			ge, err := New("", et.input, nil)
			utils.AssertNoError(t, err)
			err = ge.Setup()
			utils.AssertEqual(t, err.Error(), et.want.Error())
		}
	})

	t.Run("does not error if no setup fn defined", func(t *testing.T) {
		engine, err := New("", players.SomePlayers(), nil)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		utils.AssertNoError(t, err)
	})

	t.Run("propagates setup fn error", func(t *testing.T) {
		erroringSetupFn := func(ge GameEngine) error {
			return errors.New("Whoops")
		}
		engine, err := New("", players.SomePlayers(), erroringSetupFn)
		utils.AssertNoError(t, err)

		err = engine.Setup()
		if err == nil {
			t.Fatalf("Expected an error, but there was none")
		}
	})
}

func TestGameEngineStart(t *testing.T) {
	t.Run("only starts with legal number of players", func(t *testing.T) {

		type gameTest struct {
			testName string
			input    players.Players
			want     error
		}
		testsShouldError := []gameTest{
			{
				"too few players",
				namesToPlayers([]string{"Grace"}),
				ErrTooFewPlayers,
			},
			{
				"too many players",
				namesToPlayers([]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"}),
				ErrTooManyPlayers,
			},
		}

		for _, et := range testsShouldError {
			ge, err := New("", et.input, nil)
			utils.AssertNoError(t, err)
			err = ge.Start()
			utils.AssertEqual(t, err.Error(), et.want.Error())
		}
	})

	t.Run("unnamed for now", func(t *testing.T) {
		t.Skip("do not run TestGameStart")
		ge := gameEngineWithPlayers()

		err := ge.Start() // mock required
		if err != nil {
			t.Fatalf("Could not start game")
		}

		for _, p := range ge.Players() {
			c := p.Cards()
			numHand := len(c.Hand)
			numSeen := len(c.Seen)
			numUnseen := len(c.Unseen)
			if numHand != 3 {
				formatStr := "hand - %d\nseen - %d\nunseen - %d\n"
				t.Errorf("Expected all threes. Actual:\n" + fmt.Sprintf(formatStr, numHand, numSeen, numUnseen))
			}
		}
	})
}