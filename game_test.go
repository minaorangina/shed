package shed

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
)

func TestShedStart(t *testing.T) {
	t.Run("only starts with legal number of players", func(t *testing.T) {

		tt := []struct {
			name      string
			playerIDs []string
			err       error
		}{
			{
				"too few players",
				[]string{"Grace"},
				ErrTooFewPlayers,
			},
			{
				"too many players",
				[]string{"Ada", "Katherine", "Grace", "Hedy", "Marlyn"},
				ErrTooManyPlayers,
			},
			{
				"just right",
				[]string{"Ada", "Katherine", "Grace", "Hedy"},
				nil,
			},
		}

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				game := NewShed()
				err := game.Start(tc.playerIDs)

				utils.AssertDeepEqual(t, err, tc.err)
			})
		}
	})
}

func TestBuildMessageToPlayer(t *testing.T) {
	ge := gameEngineWithPlayers()
	ps := ge.Players()
	var opponents []Opponent
	var id string
	for _, p := range ps {
		id = p.ID()
		opponents = buildOpponents(id, ps)
		break
	}

	playerToContact, _ := ps.Find(id)
	message := buildReorgMessage(playerToContact, opponents, InitialCards{})
	expectedMessage := OutboundMessage{
		PlayerID:  playerToContact.ID(),
		Name:      playerToContact.Name(),
		Hand:      playerToContact.Cards().Hand,
		Seen:      playerToContact.Cards().Seen,
		Opponents: opponents,
		Command:   protocol.Reorg,
	}
	utils.AssertDeepEqual(t, message, expectedMessage)
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// update when GameEngine forwards messages from players

	t.Skip("do not run TestGameEngineMsgFromGame")
	ge := gameEngineWithPlayers()
	ge.Start() // mock required

	messages := []OutboundMessage{}
	want := []InboundMessage{}
	initialCards := InitialCards{}
	for _, p := range ge.Players() {
		o := buildOpponents(p.ID(), ge.Players())
		m := buildReorgMessage(p, o, initialCards)
		messages = append(messages, m)

		cards := p.Cards()

		want = append(want, InboundMessage{
			PlayerID: p.ID(),
			Hand:     cards.Hand,
			Seen:     cards.Seen,
		})
	}

	err := ge.MessagePlayers(messages)
	if err != nil {
		t.Fail()
	}
}
