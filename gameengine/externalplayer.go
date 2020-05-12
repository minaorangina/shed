package gameengine

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type conn struct {
	In  *os.File
	Out *os.File
}

// ExternalPlayer represents a player in the outside world
// An ExternalPlayer maps onto a Player in the game
type ExternalPlayer struct {
	ID   string
	Name string
	Conn *conn // tcp or command line
}

// NewExternalPlayer constructs an ExternalPlayer
func NewExternalPlayer(id, name string) ExternalPlayer {
	conn := &conn{os.Stdin, os.Stdout}
	return ExternalPlayer{
		ID:   id,
		Name: name,
		Conn: conn,
	}
}

func (ep ExternalPlayer) sendMessageAwaitReply(msg messageToPlayer) (messageFromPlayer, error) {
	response := messageFromPlayer{
		PlayerID:  msg.PlayerID,
		Command:   msg.Command,
		HandCards: msg.HandCards,
		SeenCards: msg.SeenCards,
	}

	switch msg.Command {
	case reorg:
		fmt.Println(buildReorgDisplayText(msg))
		input := make(chan bool)
		go offerCardSwitch(input, ep.Conn)
		select {
		case choice := <-input:
			fmt.Printf("Your choice: %t\n", choice)
			if choice {
				// invite to reorganise
			}
		case <-time.After(30 * time.Second):
			fmt.Println("\nTimed out: I will leave your cards as they are.")
		}
	}
	return response, nil
}

func buildReorgDisplayText(msg messageToPlayer) string {
	displayText := fmt.Sprintf("%s, here are your cards:\n\n", msg.Name)
	handText := fmt.Sprintf("In your hand, you have three cards 🤲\n")
	seenText := fmt.Sprintf("On the table, there are three more cards \n")
	unseenText := "Underneath those cards, there are three cards you can't see 🙈\n- ?\n- ?\n- ?\n"
	for _, card := range msg.SeenCards {
		seenText += "- " + card.String() + "\n"
	}
	for _, card := range msg.HandCards {
		handText += "- " + card.String() + "\n"
	}
	return displayText + handText + "\n" + seenText + "\n" + unseenText
}

func offerCardSwitch(ch chan bool, conn *conn) {
	reader := bufio.NewReader(conn.In)
	var validResponse, response bool

	for !validResponse {
		fmt.Println("You may reorganise any of your visible cards.\nWould you like to reorganise your cards? [y/n]")

		char, _, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			ch <- false
			break
		}

		switch string(char) {
		case "Y":
		case "y":
			response, validResponse = true, true
		case "N":
		case "n":
			validResponse = true
		default:
			fmt.Println("Invalid choice. Please enter \"y\" for \"yes\" or \"n\" for \"no\"")
		}
	}
	ch <- response
}
