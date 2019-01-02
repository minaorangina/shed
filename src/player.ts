import { Card, HandState } from "./types";
import { getNextHandState } from "./hand-state-machine";

export default class Player {
  name: string;
  handState: HandState;
  visibleTableCards: Card[];
  hiddenTableCards: Card[];
  handCards: Card[];

  constructor(name: string) {
    this.name = name;
  }

  setCards({
    handCards,
    visibleTableCards,
    hiddenTableCards,
    handState = HandState.A
  }) {
    this.handCards = handCards;
    this.visibleTableCards = visibleTableCards;
    this.hiddenTableCards = hiddenTableCards;
    this.handState = handState;
  }

  // updating the hand in a turn
  set hand(cards: Card[]) {
    this.handCards = this.handCards.concat(cards);

    const cardsRemaining = true;

    this.handState = getNextHandState(this.handState, {
      cardsRemaining,
      nextHandCount: this.handCards.length
    });
  }

  get hand() {
    return [...this.handCards];
  }

  getTableCards() {
    return [...this.visibleTableCards, ...this.hiddenTableCards]
  }
}
