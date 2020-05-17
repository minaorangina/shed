package gameengine

import (
	"fmt"

	"github.com/minaorangina/shed/deck"
)

func buildCardDisplayText(msg messageToPlayer) string {
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

func buildReorgDisplayText(msg messageToPlayer, visibleCards []deck.Card) string {
	displayText := "\nOk, choose the cards you wish to have in your hand\nExample: if you want cards A, C and F, type ACF (the order of the letters does not matter).\n\n"
	cardsText := ""

	for i, card := range visibleCards {
		cardsText += fmt.Sprintf("%c - %s\n", rune(upperCaseA+i), card.String())
	}

	return displayText + cardsText
}

func reorgPrompt() string {
	return "\nEnter the three cards you want in your hand: "
}
