package players

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	utils "github.com/minaorangina/shed/internal"
)

func TestOfferCardSwitch(t *testing.T) {
	t.Run("'yes' inputs", func(t *testing.T) {
		yesCases := []struct {
			input string
			want  bool
		}{
			{"y", true},
			{"Y", true},
			{"yes", true},
			{"Yes", true},
			{"YES", true},
		}

		for _, c := range yesCases {
			stdin := bytes.NewReader([]byte(c.input))
			stdout := &bytes.Buffer{}
			testConn := &conn{stdin, stdout}
			got := offerCardSwitch(testConn, time.Duration(10*time.Millisecond))

			utils.AssertEqual(t, c.want, got)
		}
	})

	t.Run("'no' inputs", func(t *testing.T) {
		noCases := []struct {
			input string
			want  bool
		}{
			{"n", false},
			{"N", false},
			{"no", false},
			{"No", false},
			{"NO", false},
		}

		for _, c := range noCases {
			stdin := bytes.NewReader([]byte(c.input))
			stdout := &bytes.Buffer{}
			testConn := &conn{stdin, stdout}
			got := offerCardSwitch(testConn, time.Duration(10*time.Millisecond))

			utils.AssertEqual(t, c.want, got)
			if !strings.Contains(stdout.String(), noChangeText) {
				t.Errorf("Expected message to contain %s, but it didn't", noChangeText)
			}
		}
	})

	t.Run("defaults to 'no' for invalid inputs", func(t *testing.T) {
		badCases := []string{
			"nah",
			"nope",
			"yup",
			"&R*$WRiyfguyfuycuiyfyiouyfuyfuyf6fW",
			"",
			" ",
			"\n",
		}

		for _, c := range badCases {
			stdin := bytes.NewReader([]byte(c))
			stdout := &bytes.Buffer{}
			testConn := &conn{stdin, stdout}
			got := offerCardSwitch(testConn, time.Duration(50*time.Millisecond))

			utils.AssertEqual(t, false, got)
			if !strings.Contains(stdout.String(), retryYesNoText) {
				t.Errorf("Got:\n%s\nShould contain:\n%s", stdout.String(), retryYesNoText)
			}
		}
	})

	t.Run("defaults to 'no' after max retries", func(t *testing.T) {
		t.Skip()
		stdin := bytes.NewReader([]byte{'%', '$', '£'})
		stdout := &bytes.Buffer{}
		testConn := &conn{stdin, stdout}
		got := offerCardSwitch(testConn, time.Duration(50*time.Millisecond))
		utils.AssertEqual(t, false, got)
		if !strings.Contains(stdout.String(), maxRetriesText) {
			t.Errorf("Got:\n%s\nShould contain:\n%s", stdout.String(), maxRetriesText)
		}
	})

	t.Run("defaults to 'no' after timeout", func(t *testing.T) {
		stdin := bytes.NewReader([]byte{})
		stdout := &bytes.Buffer{}
		testConn := &conn{stdin, stdout}
		got := offerCardSwitch(testConn, time.Duration(1*time.Millisecond))
		utils.AssertEqual(t, false, got)
		if !strings.Contains(stdout.String(), timeoutText) {
			t.Errorf("Got:\n%s\nShould contain:\n%s", stdout.String(), timeoutText)
		}
	})
}

func TestGetCardChoices(t *testing.T) {
	cases := []struct {
		input string
		want  []int
	}{
		{"ABC", []int{0, 1, 2}},
		{"abc", []int{0, 1, 2}},
		{"DEF", []int{3, 4, 5}},
		{"def", []int{3, 4, 5}},
		{"bed", []int{1, 3, 4}},
		{"DEF\nppp", []int{3, 4, 5}},
		{"nnn\nABC", []int{0, 1, 2}},
		{"\n\nfab", []int{0, 1, 5}},
		{"nnn", []int{}},
		{"\n\nneidn", []int{}},
		{"\n\n\n", []int{}},
		{"\r\n\r\n\r\n", []int{}},
	}

	for _, c := range cases {
		stdin := strings.NewReader(c.input)
		stdout := &bytes.Buffer{}
		testConn := &conn{stdin, stdout}

		got := getCardChoices(testConn, 10*time.Millisecond)

		if !reflect.DeepEqual(got, c.want) {
			utils.FailureMessage(t, c.want, got)
		}
	}
}
