package game

import (
	"testing"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/internal"
	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
	"github.com/stretchr/testify/assert"
)

func TestNewShedNewGame(t *testing.T) {
	t.Run("game with no options sets up correctly", func(t *testing.T) {
		t.Log("Given a new game")
		game := NewShed(ShedOpts{})
		playerInfo := fourPlayers()

		t.Log("When the game starts")
		err := game.Start(playerInfo)

		utils.AssertNoError(t, err)

		t.Log("Then the players are initiated correctly")
		utils.AssertTrue(t, len(game.PlayerInfo) > 1)
		utils.AssertTrue(t, len(game.ActivePlayers) == len(game.PlayerInfo))
		utils.AssertNotEmptyString(t, game.CurrentPlayer.PlayerID)

		t.Log("And the game is in the correct gameplay state")
		assert.Equal(t, gameStarted, game.gamePlay)
		assert.False(t, game.GameOver())

		t.Log("And players' cards are set correctly")
		for _, p := range game.PlayerCards {
			utils.AssertEqual(t, len(p.UnseenVisibility), 3)
		}

		for _, info := range playerInfo {
			id := info.PlayerID
			playerCards := game.PlayerCards[id]
			utils.AssertEqual(t, len(playerCards.Hand), 3)
			utils.AssertEqual(t, len(playerCards.Seen), 3)
			utils.AssertEqual(t, len(playerCards.Unseen), 3)
		}
	})
}

func TestNewShedExistingGame(t *testing.T) {
	shouldSucceed := []struct {
		name     string
		gameOpts func() ShedOpts
	}{
		{
			name: "clear deck stage: fresh",
			gameOpts: func() ShedOpts {
				d := deck.New()
				d.Shuffle()
				plrs := twoPlayers()

				cards := map[string]*PlayerCards{
					plrs[0].PlayerID: {
						Hand:   d.Deal(3),
						Seen:   d.Deal(3),
						Unseen: d.Deal(3),
					},
					plrs[1].PlayerID: {
						Hand:   d.Deal(3),
						Seen:   d.Deal(3),
						Unseen: d.Deal(3),
					},
				}

				return ShedOpts{
					Deck:            d,
					Pile:            nil,
					PlayerCards:     cards,
					Player:          plrs,
					FinishedPlayers: nil,
					CurrentPlayer:   plrs[0],
					Stage:           clearDeck,
					ExpectedCommand: protocol.PlayHand,
				}
			},
		},
		{
			name: "clear deck stage: player just played, yet to replenish",
			gameOpts: func() ShedOpts {
				d := deck.New()
				d.Shuffle()
				plrs := twoPlayers()

				cards := map[string]*PlayerCards{
					plrs[0].PlayerID: {
						Hand:   d.Deal(2),
						Seen:   d.Deal(3),
						Unseen: d.Deal(3),
					},
					plrs[1].PlayerID: {
						Hand:   d.Deal(3),
						Seen:   d.Deal(3),
						Unseen: d.Deal(3),
					},
				}

				return ShedOpts{
					Deck:            d,
					Pile:            nil,
					PlayerCards:     cards,
					Player:          plrs,
					FinishedPlayers: nil,
					CurrentPlayer:   plrs[0],
					Stage:           clearDeck,
					ExpectedCommand: protocol.ReplenishHand,
				}
			},
		},
	}

	for _, tc := range shouldSucceed {
		t.Run(tc.name, func(t *testing.T) {
			internal.ShouldNotPanic(t, func() {
				NewShed(tc.gameOpts())
			})
		})
	}
}

func TestNewShedExistingGameFailures(t *testing.T) {
	tt := []struct {
		name     string
		gameOpts func() ShedOpts
	}{}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			internal.ShouldPanic(t, func() {
				NewShed(tc.gameOpts())
			})
		})
	}
}
