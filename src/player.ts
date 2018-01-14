import { Card } from "./types";

export default class Player {
  name: string;
  tableCards: Card[];
  handCards: Card[];

  constructor({ name }) {
    this.name = name;
  }

  set startingHand(cards: Card[]) {
    this.tableCards = cards.slice(0, 6);
    let handCards = cards.slice(6, cards.length);

    for (let i = 0; i < 3; i++) {
      handCards[i].visibleToAll = true;
    }
    this.handCards = handCards;
  }

  get cards() {
    return {
      tableCards: this.tableCards,
      handCards: this.handCards
    };
  }
}
