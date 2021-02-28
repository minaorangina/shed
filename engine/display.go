package engine

import (
	"fmt"
	"io"

	"github.com/minaorangina/shed/deck"
	"github.com/minaorangina/shed/game"
	"github.com/minaorangina/shed/protocol"
)

const (
	reorgInviteText      = "You may reorganise any of your visible cards.\nWould you like to reorganise your cards? [y/n] "
	retryYesNoText       = "Invalid choice. Please enter \"y\" for \"yes\" or \"n\" for \"no\"\n"
	noChangeText         = "Ok, I will leave your cards as they are."
	timeoutText          = "\nTimed out: I will leave your cards as they are."
	maxRetriesText       = "\nMax retries exceeded: I will leave your cards as they are."
	stateOfCardsText     = "\nThanks, %s. Here is what your cards look like now:\n\n"
	startGameText        = "\nLet's start the game!"
	retryUniqueCardsText = "Please select 3 unique cards"
	retryThreeCardsText  = "You need to choose 3 cards"
	retryRangeAFText     = "Invalid entry. Please use the letter codes (A-F) to select your cards"
)

func SendText(w io.Writer, text string, a ...interface{}) {
	fmt.Fprintf(w, text, a...)
}

func buildCardDisplayText(cards game.PlayerCards) string {
	var displayText string
	handText := fmt.Sprintf("In your hand, you have three cards 🤲\n")
	seenText := fmt.Sprintf("On the table, there are three more cards \n")
	unseenText := "Underneath those cards, there are three cards you can't see 🙈\n- ?\n- ?\n- ?\n"
	for _, card := range cards.Seen {
		seenText += "- " + card.String() + "\n"
	}
	for _, card := range cards.Hand {
		handText += "- " + card.String() + "\n"
	}
	return displayText + handText + "\n" + seenText + "\n" + unseenText
}

func buildReorgDisplayText(msg protocol.OutboundMessage, visibleCards []deck.Card) string {
	displayText := "\nOk, choose the cards you wish to have in your hand\nExample: if you want cards A, C and F, type ACF (the order of the letters does not matter).\n\n"
	cardsText := ""

	for i, card := range visibleCards {
		cardsText += fmt.Sprintf("%c - %s\n", rune(upperCaseA+i), card.String())
	}

	return displayText + cardsText
}

func reorgPromptText() string {
	return "\nEnter the three cards you want in your hand: "
}
