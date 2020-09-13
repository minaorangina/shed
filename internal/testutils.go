package internal

import (
	"fmt"
	"testing"
)

// FailureMessage returns a failure message for a failed test
func FailureMessage(t *testing.T, got, want interface{}) {
	t.Helper()

	gotString := TypeToString(got)
	wantString := TypeToString(want)
	t.Errorf("\nGot: %s\nwant: %s", gotString, wantString)
}

// TableFailureMessage returns a failure message for a failed test, including the name of the test
func TableFailureMessage(t *testing.T, testName, got, want interface{}) {
	t.Helper()

	actualString := TypeToString(got)
	expectedString := TypeToString(want)
	t.Errorf("%s\nGot: %s\nWant: %s", testName, actualString, expectedString)
}

// TypeToString returns the string representation of a non-string type
func TypeToString(obj interface{}) string {
	return fmt.Sprintf("%+v", obj)
}

// AssertNoError checks for the non-existence of an error
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

// AssertEqual checks that the values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()

	if got != want {
		FailureMessage(t, got, want)
	}
}

func AssertStringEquality(t *testing.T, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("got %s, want %s", got, want)
	}
}
