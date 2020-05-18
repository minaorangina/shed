package gameengine

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
	upperCaseA = 65
	upperCaseD = 65
	upperCaseF = 70
	retries    = 3
)

func (ep ExternalPlayer) handleReorg(msg messageToPlayer) (messageFromPlayer, error) {
	response := messageFromPlayer{
		PlayerID: msg.PlayerID,
		Command:  msg.Command,
		Hand:     msg.Hand,
		Seen:     msg.Seen,
	}

	fmt.Println(buildCardDisplayText(msg))

	if shouldReorganise := offerCardSwitch(ep.Conn); shouldReorganise {
		response = reorganiseCards(ep.Conn, msg)
	}

	return response, nil
}

func offerCardSwitch(conn *conn) bool {
	input := make(chan bool)
	go func(ch chan bool) {
		reader := bufio.NewReader(conn.In)
		var validResponse, response bool

		for !validResponse {
			fmt.Printf("You may reorganise any of your visible cards.\nWould you like to reorganise your cards? [y/n] ")

			char, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Println(err.Error())
				fmt.Printf("Error: %s\n", err.Error())
				ch <- false
				break
			}
			char = strings.TrimSpace(char)

			switch char {
			case "Y":
			case "y":
				response, validResponse = true, true
			case "N":
			case "n":
				validResponse = true
			default:
				fmt.Printf("Invalid choice %s. Please enter \"y\" for \"yes\" or \"n\" for \"no\"\n", char)
			}
		}
		ch <- response
	}(input)

	select {
	case choice := <-input:
		if choice {
			return true
		}
		fmt.Println("Ok, I will leave your cards as they are.")
		return false

	case <-time.After(3 * time.Second):
		fmt.Println("\nTimed out: I will leave your cards as they are.")
		return false
	}
}

func reorganiseCards(conn *conn, msg messageToPlayer) messageFromPlayer {
	defaultResponse := messageFromPlayer{
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

	fmt.Printf(buildReorgDisplayText(msg, allVisibleCards))

	ch := make(chan []int)
	go getCardChoices(ch, conn)

	select {
	case choices := <-ch:
		if len(choices) == 3 {
			newHand, newSeen := choicesToCards(allVisibleCards, choices)

			return messageFromPlayer{
				PlayerID: msg.PlayerID,
				Command:  msg.Command,
				Hand:     newHand,
				Seen:     newSeen,
			}
		}
		return defaultResponse

	case <-time.After(1 * time.Minute):
		fmt.Println("Ok, I'll leave your cards as they are")
		return defaultResponse
	}
}

func getCardChoices(ch chan []int, conn *conn) {
	var validResponse bool
	var response []int
	reader := bufio.NewReader(conn.In)
	retriesLeft := retries

	for retriesLeft > 0 && !validResponse {
		fmt.Printf(reorgPrompt())

		entryBytes, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		}
		entry := strings.Replace(string(entryBytes), " ", "", -1)

		if len(entry) != 3 {
			fmt.Println("You need to choose 3 cards")
			retriesLeft--
			continue
		}
		if !charsUnique(entry) {
			fmt.Println("Please select 3 unique cards")
			retriesLeft--
			continue
		}
		entry = strings.ToUpper(entry)
		if !charsInRange(entry, upperCaseA, upperCaseF) {
			fmt.Println("Invalid entry. Please use the letter codes (A-F) to select your cards")
			retriesLeft--
			continue
		}

		validResponse = true
		response = charsToSortedCardIndex(entry)
	}

	ch <- response
}

func charsInRange(chars string, lower, upper int) bool {
	for _, char := range chars {
		if int(char) < lower || int(char) > upper {
			return false
		}
	}
	return true
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

func charsUnique(s string) bool {
	seen := map[string]bool{}
	for _, c := range s {
		if _, ok := seen[string(c)]; ok {
			return false
		}
		seen[string(c)] = true
	}
	return true
}
