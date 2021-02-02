package shed

import (
	"bufio"
	"sort"
	"strings"
	"time"

	"github.com/minaorangina/shed/deck"
)

var (
	upperCaseA   = 65
	upperCaseD   = 65
	upperCaseF   = 70
	retries      = 3
	offerTimeout = 30 * time.Second
	reorgTimeout = 1 * time.Minute
)

func offerCardSwitch(conn *conn, timeout time.Duration) bool {
	input := make(chan bool)
	go func(ch chan bool) {
		reader := bufio.NewScanner(conn.In)

		var validResponse, response bool
		for !validResponse {
			SendText(conn.Out, reorgInviteText)
			reader.Scan()
			char := strings.TrimSpace(reader.Text())

			switch char {
			case "Y":
				fallthrough
			case "y":
				fallthrough
			case "YES":
				fallthrough
			case "yes":
				fallthrough
			case "Yes":
				response, validResponse = true, true
			case "N":
				fallthrough
			case "n":
				fallthrough
			case "no":
				fallthrough
			case "No":
				fallthrough
			case "NO":
				validResponse = true
			default:
				SendText(conn.Out, retryYesNoText)
			}
		}
		ch <- response
	}(input)

	select {
	case choice := <-input:
		if choice {
			return true
		}
		SendText(conn.Out, noChangeText)
		return false

	case <-time.After(timeout):
		SendText(conn.Out, timeoutText)
		return false
	}
}

func reorganiseCards(conn *conn, msg OutboundMessage) InboundMessage {
	defaultResponse := InboundMessage{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Decision: []int{0, 1, 2},
	}

	allVisibleCards := []deck.Card{}
	for _, c := range msg.Hand {
		allVisibleCards = append(allVisibleCards, c)
	}
	for _, c := range msg.Seen {
		allVisibleCards = append(allVisibleCards, c)
	}

	SendText(conn.Out, buildReorgDisplayText(msg, allVisibleCards))

	choices := getCardChoices(conn, reorgTimeout)
	if len(choices) == 3 {
		newHand, newSeen := choicesToCards(allVisibleCards, choices)
		playerCards := PlayerCards{
			Seen: newSeen,
			Hand: newHand,
		}
		SendText(conn.Out, stateOfCardsText, msg.Name)
		SendText(conn.Out, buildCardDisplayText(playerCards))
		SendText(conn.Out, startGameText)

		return InboundMessage{
			PlayerID: msg.PlayerID,
			Command:  msg.Command,
			Decision: choices,
		}
	}

	SendText(conn.Out, noChangeText)
	return defaultResponse
}

func getCardChoices(conn *conn, timeout time.Duration) []int {
	input := make(chan []int)

	go func(ch chan []int) {
		var validResponse bool
		response := []int{}
		retriesLeft := retries

		reader := bufio.NewScanner(conn.In)

		for retriesLeft > 0 && !validResponse {
			SendText(conn.Out, reorgPromptText())
			reader.Scan()

			entry := strings.Replace(reader.Text(), " ", "", -1)
			if len(entry) != 3 {
				SendText(conn.Out, retryThreeCardsText)
				retriesLeft--
				continue
			}

			if !charsUnique(entry) {
				SendText(conn.Out, retryUniqueCardsText)
				retriesLeft--
				continue
			}

			entry = strings.ToUpper(entry)

			if !charsInRange(entry, upperCaseA, upperCaseF) {
				SendText(conn.Out, retryRangeAFText)
				retriesLeft--
				continue
			}

			validResponse = true
			response = charsToSortedCardIndex(entry)
		}

		ch <- response
		close(ch) // necessary?
	}(input)

	select {
	case choices := <-input:
		return choices
	case <-time.After(timeout):
		return []int{}
	}
}

func charsToSortedCardIndex(chars string) []int {
	indices := []int{}
	for _, char := range chars {
		indices = append(indices, int(char)-upperCaseA)
	}
	sort.Ints(indices)
	return indices
}

func choicesToCards(allCards []deck.Card, choices []int) ([]deck.Card, []deck.Card) {
	newHand := []deck.Card{}
	newSeen := []deck.Card{}

	for i, choiceIdx := 0, 0; i < len(allCards); i++ {
		if choiceIdx < len(choices) && i == choices[choiceIdx] {
			newHand = append(newHand, allCards[i])
			choiceIdx++
		} else {
			newSeen = append(newSeen, allCards[i])
		}
	}
	return newHand, newSeen
}
