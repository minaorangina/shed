package gameengine

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestGameEngine(t *testing.T) {
	type engineTest struct {
		testName string
		input    []string
		expected error
	}

	testsShouldError := []engineTest{
		{
			"too few players",
			[]string{"Grace"},
			errors.New("Could not construct GameEngine: minimum of 2 players required (supplied 1)"),
		},
		{
			"too many players",
			[]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"},
			errors.New("Could not construct GameEngine: maximum of 4 players required (supplied 5)"),
		},
	}

	for _, et := range testsShouldError {
		_, err := New(et.input)
		if err == nil {
			t.Errorf(utils.TableFailureMessage(et.testName, strings.Join(et.input, ","), et.expected.Error()))
		}
	}

	// Construct a GameEngine
	playerNames := []string{"Ada", "Katherine"}
	engine, err := New(playerNames)

	if err != nil {
		t.Fail()
	}
	expectedEngine := GameEngine{
		playerNames: playerNames,
	}
	if !reflect.DeepEqual(expectedEngine, engine) {
		t.Errorf("\nExpected: %+v\nActual: %+v", expectedEngine, engine)
	}
}
