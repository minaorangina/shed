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
          name: Deck.generateName(value, suit)
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

  order() {
    this.cards = this.cards.sort(this.orderByProperty(["suit", "value"]));
  }

  private orderByProperty(props: Array<any>) {
    const currentArg = props[0];
    const restArgs = props.slice(1);
    const self = this;
    return function(a, b) {
      let equality;
      if (a[currentArg] < b[currentArg]) {
        equality = -1;
      } else if (a[currentArg] > b[currentArg]) {
        equality = 1;
      } else {
        equality = 0;
      }
      // if of equal value, and there are still properties to compare by
      // recall the function, comparing the same two values,
      // but compare with the next property
      if (equality === 0 && props.length > 1) {
        return self.orderByProperty.call(self, restArgs)(a, b);
      }
      return equality;
    };
  }
}

export default Deck;