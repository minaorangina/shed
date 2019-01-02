import Deck from "./deck";
import { Card } from "./types";
import Player from "./player";

export default class GameEngine {
  players: Player[];
  unplayed: Card[];

  constructor(playerNames: string[]) {
    let numPlayers = playerNames.length;
    if (numPlayers < 2) {
      throw new RangeError("You need at least 2 players to play :)");
    }

    const players = playerNames.map(name => {
      return new Player(name);
    });

    this.dealCards(players);
    this.players = players;
  }

  dealCards(players: Player[]) {
    const deck = new Deck();
    const cardsPerPlayer = 9;
    this.unplayed = deck.cards.slice(cardsPerPlayer * players.length);

    players.forEach((player, i) => {
      const startIndex = cardsPerPlayer * i;
      const endIndex = startIndex + cardsPerPlayer;
      const cards = deck.cards.slice(startIndex, endIndex);

      const handCards = cards.slice(0, 3);
      const visibleTableCards = cards.slice(3, 6);
      visibleTableCards.forEach(card => (card.visibleToAll = true));
      const hiddenTableCards = cards.slice(6, 9);

      player.setCards({
        handCards,
        visibleTableCards,
        hiddenTableCards
      });
    });
  }

  getCardsInPlay() {
    return this.players.map(p => {
      return {
        hand: p.hand,
        tableCards: p.getTableCards()
      };
    });
  }
}
