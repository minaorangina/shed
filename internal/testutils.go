package internal

import "fmt"

// FailureMessage returns a failure message for a failed test
func FailureMessage(expected, actual string) string {
	return fmt.Sprintf("\nExpected: %s\nActual: %s", expected, actual)
}

// TableFailureMessage returns a failure message for a failed test, including the name of the test
func TableFailureMessage(testName, expected, actual string) string {
	return fmt.Sprintf("%s\nExpected: %s\nActual: %s", testName, expected, actual)
}

// TypeToString returns the string representation of a non-string type
func TypeToString(obj interface{}) string {
	return fmt.Sprintf("%+v", obj)
}
