package gameengine

import (
	"bytes"
	"testing"
)

func TestSendText(t *testing.T) {
	t.Run("send simple text", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		want := "Hello"
		SendText(buffer, want)

		got := buffer.String()

		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("send formatted text", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		want := "Hello, human"
		format := "Hello, %s"
		SendText(buffer, format, "human")

		got := buffer.String()

		if want != got {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}
