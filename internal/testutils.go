package internal

import (
	"fmt"
	"testing"
)

// FailureMessage returns a failure message for a failed test
func FailureMessage(t *testing.T, expected, actual interface{}) {
	expectedString := TypeToString(expected)
	actualString := TypeToString(actual)
	t.Errorf("\nExpected: %s\nActual: %s", expectedString, actualString)
}

// TableFailureMessage returns a failure message for a failed test, including the name of the test
func TableFailureMessage(t *testing.T, testName, expected, actual interface{}) {
	expectedString := TypeToString(expected)
	actualString := TypeToString(actual)
	t.Errorf("%s\nExpected: %s\nActual: %s", testName, expectedString, actualString)
}

// TypeToString returns the string representation of a non-string type
func TypeToString(obj interface{}) string {
	return fmt.Sprintf("%+v", obj)
}
