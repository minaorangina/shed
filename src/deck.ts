import { Card, Suit } from "./types";

class Deck {
  cards: Card[];

  constructor() {
    this.cards = [];
    for (let suit in Suit) {
      let value = 1;
      do {
        this.cards.push({
          value,
          suit,
          name: Deck.generateName(value, suit),
          visibleToAll: false,
          inPlay: true
        });
        value++;
      } while (value <= 13);
    }
    this.shuffle();
  }

  static generateName(value, suit) {
    let cardVal = value;
    if (value === 1) {
      cardVal = "Ace";
    }
    if (value === 11) {
      cardVal = "Jack";
    }
    if (value === 12) {
      cardVal = "Queen";
    }
    if (value === 13) {
      cardVal = "King";
    }
    return `${cardVal} of ${suit}`;
  }

  shuffle() {
    let toShuffle = this.cards.length,
      tmp,
      randCard;

    // While there remain elements to shuffle…
    while (toShuffle) {
      // Pick a remaining element…
      randCard = Math.floor(Math.random() * toShuffle--);

      // And swap it with the current element.
      tmp = this.cards[toShuffle];
      this.cards[toShuffle] = this.cards[randCard];
      this.cards[randCard] = tmp;
    }
  }
}

export default Deck;
