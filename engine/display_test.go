package engine

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestSendText(t *testing.T) {
	t.Run("send simple text", func(t *testing.T) {
		buffer := NewTestBuffer()
		want := "Hello"
		SendText(buffer, want)

		got := buffer.String()

		utils.AssertStringEquality(t, got, want)
	})

	t.Run("send formatted text", func(t *testing.T) {
		buffer := NewTestBuffer()
		want := "Hello, human"
		format := "Hello, %s"
		SendText(buffer, format, "human")

		got := buffer.String()

		utils.AssertStringEquality(t, got, want)
	})
}
