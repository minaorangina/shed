package internal

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
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

// AssertErrored checks for the existence of an error
func AssertErrored(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
}

// AssertEqual checks that the values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()

	if got != want {
		FailureMessage(t, got, want)
	}
}

// AssertDeepEqual checks that the values are equal
func AssertDeepEqual(t *testing.T, got, want interface{}) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		FailureMessage(t, got, want)
	}
}

// AssertNotDeepEqual checks that the values are not equal
func AssertNotDeepEqual(t *testing.T, a, b interface{}) {
	t.Helper()

	if reflect.DeepEqual(a, b) {
		t.Error("unexpected deep equality")
	}
}

// AssertEqualToOneOf checks that at least one of the values are equal
func AssertEqualToOneOf(t *testing.T, got interface{}, want ...interface{}) {
	t.Helper()
	for _, w := range want {
		if got == w {
			return
		}
	}
	FailureMessage(t, got, want)
}

func AssertStringEquality(t *testing.T, got, want string) {
	t.Helper()
	if want != got {
		t.Errorf("got %s, want %s", got, want)
	}
}

// AssertTrue checks that the value is true
func AssertTrue(t *testing.T, got bool) {
	t.Helper()

	if got != true {
		t.Error("Expected to be true, but it wasn't")
	}
}

// AssertNotNil checks that the value is not nil
func AssertNotNil(t *testing.T, got interface{}) {
	t.Helper()

	if got == nil {
		t.Error("Value is unexpectedly nil")
	}
}

// AssertNotEmptyString checks the string is not the empty string
func AssertNotEmptyString(t *testing.T, got string) {
	t.Helper()

	if got == "" {
		t.Error("unexpected empty string")
	}
}

// AssertContains checks if the string contains the specified substring
func AssertContains(t *testing.T, got, substring string) {
	t.Helper()

	if !strings.Contains(got, substring) {
		t.Errorf("expected '%s' to contain %s (but it didn't)", got, substring)
	}
}

func Within(t *testing.T, d time.Duration, assert func()) {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		assert()
		done <- struct{}{}
	}()

	select {
	case <-time.After(d):
		t.Error("timed out")
	case <-done:
	}
}

// ShouldPanic checks if the given function call results in a panic
func ShouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected to panic, but it didn't")
		}
	}()

	f()
}
