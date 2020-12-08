package shed

import (
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/protocol"
)

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
