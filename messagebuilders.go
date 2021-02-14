package shed

import (
	"fmt"

	"github.com/minaorangina/shed/protocol"
)

func (s *shed) buildBaseMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:    playerID,
		CurrentTurn: s.CurrentPlayer,
		Hand:        s.PlayerCards[playerID].Hand,
		Seen:        s.PlayerCards[playerID].Seen,
		Pile:        s.Pile,
		DeckCount:   len(s.Deck),
	}
}

func (s *shed) buildOpponents(playerID string) []Opponent {
	opponents := []Opponent{}

	for _, p := range s.PlayerInfo {
		if p.PlayerID != playerID {
			opponents = append(opponents, Opponent{
				PlayerID: p.PlayerID,
				Name:     p.Name,
				Seen:     s.PlayerCards[p.PlayerID].Seen,
			})
		}
	}

	return opponents
}

func (s *shed) buildSkipTurnMessage(playerID string) OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.SkipTurn
	msg.Message = fmt.Sprintf("%s skips a turn!", s.CurrentPlayer.Name)
	msg.Opponents = s.buildOpponents(s.CurrentPlayer.PlayerID)

	return msg
}

func (s *shed) buildSkipTurnMessages(currentPlayerCmd protocol.Cmd) []OutboundMessage {
	currentPlayerMsg := s.buildBaseMessage(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.Command = protocol.SkipTurn
	currentPlayerMsg.Message = "You skip a turn!"
	currentPlayerMsg.Opponents = s.buildOpponents(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.ShouldRespond = true

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, info := range s.PlayerInfo {
		if info.PlayerID != s.CurrentPlayer.PlayerID {
			toSend = append(toSend, s.buildSkipTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildTurnMessage(playerID string) OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.Turn
	msg.Message = fmt.Sprintf("It's %s's turn!", s.CurrentPlayer.Name)
	msg.Opponents = s.buildOpponents(playerID)

	return msg
}

func (s *shed) buildTurnMessages(currentPlayerCmd protocol.Cmd, moves []int) []OutboundMessage {
	var toPlay string
	switch currentPlayerCmd {
	case protocol.PlayHand:
		toPlay = "hand"
	case protocol.PlaySeen:
		toPlay = "face-up"
	case protocol.PlayUnseen:
		toPlay = "face-down"
	}

	displayMsg := "It's your turn!"
	if toPlay != "" {
		displayMsg += fmt.Sprintf(" Play from your %s cards.", toPlay)
	}

	currentPlayerMsg := s.buildBaseMessage(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.Command = currentPlayerCmd
	currentPlayerMsg.Message = displayMsg
	currentPlayerMsg.Opponents = s.buildOpponents(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.Moves = moves
	currentPlayerMsg.ShouldRespond = true

	toSend := []OutboundMessage{currentPlayerMsg}
	for _, info := range s.PlayerInfo {
		if info.PlayerID != s.CurrentPlayer.PlayerID {
			toSend = append(toSend, s.buildTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildEndOfTurnMessage(playerID string) OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.EndOfTurn

	return msg
}

func (s *shed) buildEndOfTurnMessages(currentPlayerCommand protocol.Cmd) []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildEndOfTurnMessage(info.PlayerID)
		if info.PlayerID == s.CurrentPlayer.PlayerID {
			msg.Command = currentPlayerCommand
			msg.ShouldRespond = true
		}
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildPlayerFinishedMessage(playerID string) OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.PlayerFinished
	msg.Message = fmt.Sprintf("%s has finished!", s.CurrentPlayer.Name)
	msg.FinishedPlayers = s.FinishedPlayers

	return msg
}

func (s *shed) buildPlayerFinishedMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildPlayerFinishedMessage(info.PlayerID)
		if info.PlayerID == s.CurrentPlayer.PlayerID {
			msg.ShouldRespond = true
			msg.Message = "You've finished!"
		}
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildGameOverMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildBaseMessage(info.PlayerID)
		msg.Command = protocol.GameOver
		msg.Message = "Game over!"
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildBurnMessage(playerID string) OutboundMessage {
	return OutboundMessage{
		PlayerID:        playerID,
		Command:         protocol.Burn,
		CurrentTurn:     s.CurrentPlayer,
		Hand:            s.PlayerCards[playerID].Hand,
		Seen:            s.PlayerCards[playerID].Seen,
		Pile:            s.Pile,
		FinishedPlayers: s.FinishedPlayers,
		Message:         fmt.Sprintf("Burn for %s!", s.CurrentPlayer.Name),
	}
}

func (s *shed) buildBurnMessages() []OutboundMessage {
	toSend := []OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildBurnMessage(info.PlayerID)
		if info.PlayerID == s.CurrentPlayer.PlayerID {
			msg.ShouldRespond = true
			msg.Message = "Burn!"
		}
		toSend = append(toSend, msg)
	}

	return toSend
}
