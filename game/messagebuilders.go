package game

import (
	"fmt"

	"github.com/minaorangina/shed/protocol"
)

func (s *shed) buildBaseMessage(playerID string) protocol.OutboundMessage {
	playerCards := s.PlayerCards[playerID]
	publicUnseen := s.mapUnseenToPublicUnseen(playerID)

	return protocol.OutboundMessage{
		PlayerID:    playerID,
		CurrentTurn: s.CurrentPlayer,
		Hand:        playerCards.Hand,
		Seen:        playerCards.Seen,
		Unseen:      publicUnseen,
		Pile:        s.Pile,
		DeckCount:   len(s.Deck),
	}
}

func (s *shed) buildOpponents(playerID string) []protocol.Opponent {
	opponents := []protocol.Opponent{}

	for _, p := range s.PlayerInfo {
		if p.PlayerID != playerID {
			opponents = append(opponents, protocol.Opponent{
				PlayerID: p.PlayerID,
				Name:     p.Name,
				Seen:     s.PlayerCards[p.PlayerID].Seen,
			})
		}
	}

	return opponents
}

func (s *shed) buildReorgMessages() []protocol.OutboundMessage {
	msgs := []protocol.OutboundMessage{}

	for _, info := range s.PlayerInfo {
		m := s.buildBaseMessage(info.PlayerID)
		m.Command = protocol.Reorg
		m.Message = "Choose the cards you want in your hand. The remaining cards will become your face-up cards."
		m.ShouldRespond = true

		msgs = append(msgs, m)
	}

	return msgs
}

func (s *shed) buildReplenishHandMessages() []protocol.OutboundMessage {
	msgs := []protocol.OutboundMessage{}
	currentPlayerMsg := s.buildBaseMessage(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.Command = protocol.ReplenishHand
	currentPlayerMsg.ShouldRespond = true

	msgs = append(msgs, currentPlayerMsg)

	for _, info := range s.PlayerInfo {
		if info.PlayerID != s.CurrentPlayer.PlayerID {
			msgs = append(msgs, s.buildEndOfTurnMessage(info.PlayerID))
		}
	}

	return msgs
}

func (s *shed) buildSkipTurnMessage(playerID string) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.SkipTurn
	msg.Message = fmt.Sprintf("%s skips a turn!", s.CurrentPlayer.Name)
	msg.Opponents = s.buildOpponents(s.CurrentPlayer.PlayerID)

	return msg
}

func (s *shed) buildSkipTurnMessages(currentPlayerCmd protocol.Cmd) []protocol.OutboundMessage {
	currentPlayerMsg := s.buildBaseMessage(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.Command = protocol.SkipTurn
	currentPlayerMsg.Message = "You skip a turn!"
	currentPlayerMsg.Opponents = s.buildOpponents(s.CurrentPlayer.PlayerID)
	currentPlayerMsg.ShouldRespond = true

	toSend := []protocol.OutboundMessage{currentPlayerMsg}
	for _, info := range s.PlayerInfo {
		if info.PlayerID != s.CurrentPlayer.PlayerID {
			toSend = append(toSend, s.buildSkipTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildTurnMessage(playerID string) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.Turn
	msg.Message = fmt.Sprintf("It's %s's turn!", s.CurrentPlayer.Name)
	msg.Opponents = s.buildOpponents(playerID)

	return msg
}

func (s *shed) buildTurnMessages(currentPlayerCmd protocol.Cmd, moves []int) []protocol.OutboundMessage {
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

	toSend := []protocol.OutboundMessage{currentPlayerMsg}
	for _, info := range s.PlayerInfo {
		if info.PlayerID != s.CurrentPlayer.PlayerID {
			toSend = append(toSend, s.buildTurnMessage(info.PlayerID))
		}
	}

	return toSend
}

func (s *shed) buildEndOfTurnMessage(playerID string) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.EndOfTurn

	return msg
}

func (s *shed) buildEndOfTurnMessages(currentPlayerCommand protocol.Cmd) []protocol.OutboundMessage {
	toSend := []protocol.OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildEndOfTurnMessage(info.PlayerID)

		if info.PlayerID == s.CurrentPlayer.PlayerID {
			msg.Command = currentPlayerCommand
			msg.ShouldRespond = true

			if currentPlayerCommand == protocol.UnseenSuccess {
				msg.Message = "Good move!"
			}
			if currentPlayerCommand == protocol.UnseenFailure {
				msg.Message = "Bad luck!"
			}
		} else {
			if currentPlayerCommand == protocol.UnseenSuccess {
				msg.Message = fmt.Sprintf("%s's card choice succeeds!", s.CurrentPlayer.Name)
			}
			if currentPlayerCommand == protocol.UnseenFailure {
				msg.Message = fmt.Sprintf("%s's card choice fails!", s.CurrentPlayer.Name)
			}
		}

		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildPlayerFinishedMessage(playerID string) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.PlayerFinished
	msg.Message = fmt.Sprintf("%s has finished!", s.CurrentPlayer.Name)
	msg.FinishedPlayers = s.FinishedPlayers

	return msg
}

func (s *shed) buildPlayerFinishedMessages() []protocol.OutboundMessage {
	toSend := []protocol.OutboundMessage{}
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

func (s *shed) buildGameOverMessages() []protocol.OutboundMessage {
	toSend := []protocol.OutboundMessage{}
	for _, info := range s.PlayerInfo {
		msg := s.buildBaseMessage(info.PlayerID)
		msg.Command = protocol.GameOver

		var result string
		for position, finshedPlayer := range s.FinishedPlayers {
			if info.PlayerID == finshedPlayer.PlayerID {
				switch position {
				case 0: // first place
					result = "won!"
				case len(s.FinishedPlayers) - 1: // last place
					result = "lost :("
				default:
					result = "didn't win, but most importantly, you didn't lose :)"
				}
			}
		}

		msg.Message = fmt.Sprintf("Game over! You %s", result)
		toSend = append(toSend, msg)
	}

	return toSend
}

func (s *shed) buildBurnMessage(playerID string) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.Burn
	msg.Message = fmt.Sprintf("Burn for %s!", s.CurrentPlayer.Name)

	return msg
}

func (s *shed) buildBurnMessages() []protocol.OutboundMessage {
	toSend := []protocol.OutboundMessage{}
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

func (s *shed) buildErrorMessages(err error) []protocol.OutboundMessage {
	msgs := []protocol.OutboundMessage{}
	for _, p := range s.PlayerInfo {
		msgs = append(msgs, s.buildErrorMessage(p.PlayerID, err))
	}

	return msgs
}

func (s *shed) buildErrorMessage(playerID string, err error) protocol.OutboundMessage {
	msg := s.buildBaseMessage(playerID)
	msg.Command = protocol.Error
	msg.Message = fmt.Sprintf("game error: %q", err.Error())
	msg.ShouldRespond = s.AwaitingResponse() != protocol.Null && s.CurrentPlayer.PlayerID == playerID

	return msg
}

func BuildGameHasStartedMessage(playerID, name string) protocol.OutboundMessage {
	return protocol.OutboundMessage{
		PlayerID: playerID,
		Name:     name,
		Message:  fmt.Sprintf("Game is starting!"),
		Command:  protocol.HasStarted,
	}
}

func BuildNewJoinerMessage(playerID, name, joinerPlayerID, joinerName string) protocol.OutboundMessage {
	return protocol.OutboundMessage{
		PlayerID: playerID,
		Name:     name,
		Joiner:   protocol.PlayerInfo{Name: joinerName, PlayerID: joinerPlayerID},
		Message:  fmt.Sprintf("%s has joined the game!", joinerName),
		Command:  protocol.NewJoiner,
	}
}
