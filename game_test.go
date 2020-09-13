package shed

import (
	"reflect"
	"testing"

	utils "github.com/minaorangina/shed/internal"
	"github.com/minaorangina/shed/players"
	"github.com/minaorangina/shed/protocol"
)

func TestBuildMessageToPlayer(t *testing.T) {
	ge := gameEngineWithPlayers()
	var opponents []players.Opponent
	var id string
	for _, p := range ge.Players() {
		id = p.ID
		opponents = buildOpponents(id, ge.Players())
		break
	}

	playerToContact, _ := ge.Players().Find(id)
	message := buildReorgMessage(playerToContact, opponents, players.InitialCards{}, "Let the games begin!")
	expectedMessage := players.OutboundMessage{
		Message:   "Let the games begin!",
		PlayerID:  playerToContact.ID,
		Name:      playerToContact.Name,
		Hand:      playerToContact.Cards().Hand,
		Seen:      playerToContact.Cards().Seen,
		Opponents: opponents,
		Command:   protocol.Reorg,
	}
	if !reflect.DeepEqual(expectedMessage, message) {
		utils.FailureMessage(t, message, expectedMessage)
	}
}

func TestGameEngineMsgFromGame(t *testing.T) {
	// Game Engine receives from messages to send to players
	// and returns response
	t.Skip("do not run TestGameEngineMsgFromGame")
	ge := gameEngineWithPlayers()
	ge.Start() // mock required

	messages := []players.OutboundMessage{}
	want := []players.InboundMessage{}
	initialCards := players.InitialCards{}
	for _, p := range ge.Players() {
		o := buildOpponents(p.ID, ge.Players())
		m := buildReorgMessage(p, o, initialCards, "Rearrange your initial cards")
		messages = append(messages, m)

		want = append(want, players.InboundMessage{
			PlayerID: p.ID,
			Hand:     p.Hand,
			Seen:     p.Seen,
		})
	}

	got, err := ge.MessagePlayers(messages)
	if err != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(got, want) { // TODO: fix slice ordering
		utils.FailureMessage(t, got, want)
	}
}
