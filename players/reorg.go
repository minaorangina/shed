package players

import (
	"bufio"
	"fmt"
	"io"
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

func offerCardSwitch(conn *conn) bool {
	input := make(chan bool)
	go func(ch chan bool) {
		reader := bufio.NewReader(conn.In)
		var validResponse, response bool
		for !validResponse {
			SendText(conn.Out, reorgInviteText)

			char, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				// should go to a logger eventually
				fmt.Println(err.Error())
				fmt.Printf("Error: %s\n", err.Error())
				ch <- false
				break
			}
			char = strings.TrimSpace(char)

			switch char {
			case "Y":
				fallthrough
			case "y":
				response, validResponse = true, true
			case "N":
				fallthrough
			case "n":
				validResponse = true
			default:
				SendText(conn.Out, retryYesNoText, char)
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

	case <-time.After(offerTimeout):
		SendText(conn.Out, timeoutText)
		return false
	}
}

func reorganiseCards(conn *conn, msg OutboundMessage) InboundMessage {
	defaultResponse := InboundMessage{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Hand:     msg.Hand,
		Seen:     msg.Seen,
	}

	allVisibleCards := []deck.Card{}
	for _, c := range msg.Hand {
		allVisibleCards = append(allVisibleCards, c)
	}
	for _, c := range msg.Seen {
		allVisibleCards = append(allVisibleCards, c)
	}

	SendText(conn.Out, buildReorgDisplayText(msg, allVisibleCards))

	ch := make(chan []int)
	go getCardChoices(ch, conn)

	select {
	case choices := <-ch:
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
				Hand:     newHand,
				Seen:     newSeen,
			}
		}

		SendText(conn.Out, noChangeText)
		return defaultResponse

	case <-time.After(reorgTimeout):
		SendText(conn.Out, noChangeText)
		return defaultResponse
	}
}

func getCardChoices(ch chan []int, conn *conn) {
	var validResponse bool
	var response []int
	reader := bufio.NewReader(conn.In)
	retriesLeft := retries

	for retriesLeft > 0 && !validResponse {
		SendText(conn.Out, reorgPromptText())

		entryBytes, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		entry := strings.Replace(string(entryBytes), " ", "", -1)

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
