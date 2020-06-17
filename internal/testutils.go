package internal

import (
	"fmt"
	"testing"
)

// FailureMessage returns a failure message for a failed test
func FailureMessage(t *testing.T, want, got interface{}) {
	expectedString := TypeToString(want)
	actualString := TypeToString(got)
	t.Errorf("\nWant: %s\nGot: %s", expectedString, actualString)
}

// TableFailureMessage returns a failure message for a failed test, including the name of the test
func TableFailureMessage(t *testing.T, testName, want, got interface{}) {
	expectedString := TypeToString(want)
	actualString := TypeToString(got)
	t.Errorf("%s\nWant: %s\nGot: %s", testName, expectedString, actualString)
}

// TypeToString returns the string representation of a non-string type
func TypeToString(obj interface{}) string {
	return fmt.Sprintf("%+v", obj)
}

// AssertNoError checks for the non-existence of an error
func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

// AssertEqual checks that the values are equal
func AssertEqual(t *testing.T, want, got interface{}) {
	if want != got {
		FailureMessage(t, want, got)
	}
}
