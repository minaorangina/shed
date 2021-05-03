package game

import (
	"testing"

	"github.com/minaorangina/shed/protocol"
	"github.com/stretchr/testify/require"
)

func TestBuildBaseMessage(t *testing.T) {
	shed := NewShed(ShedOpts{})
	players := []protocol.PlayerInfo{
		{
			PlayerID: "1",
			Name:     "Helga",
		},
		{
			PlayerID: "2",
			Name:     "Helena",
		},
	}
	err := shed.Start(players)
	require.NoError(t, err)

	msgs, err := shed.Next()
	require.NoError(t, err)
	require.Equal(t, 2, len(msgs))

	for _, m := range msgs {
		require.NotEmpty(t, m.PlayerID)
		require.NotEmpty(t, m.Hand)
		require.NotEmpty(t, m.Seen)
		require.NotEmpty(t, m.Unseen)
		require.NotEmpty(t, m.DeckCount)
		require.NotEmpty(t, m.CurrentTurn.Name)
		require.NotEmpty(t, m.CurrentTurn.PlayerID)
		require.NotEmpty(t, m.NextTurn.Name)
		require.NotEmpty(t, m.NextTurn.PlayerID)

		require.NotEqual(t, m.NextTurn, m.CurrentTurn)
	}
}
