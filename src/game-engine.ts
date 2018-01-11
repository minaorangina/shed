import Deck from "./deck";
import { Card } from "./types";
import Player from "./player";

export default class GameEngine {
  game: Game;

  constructor(players = []) {
    let numPlayers = players.length;
    if (numPlayers < 2) {
      throw new RangeError("You need at least 2 players to play :)");
    }
    this.initGame(players);
  }

  initGame(players: Array<any>) {
    this.game = new Game(players);
  }
}

export class Game {
  players: Player[];
  unplayed: Card[];

  constructor(players) {
    this.dealCards(players);
  }

  dealCards(players: Array<any>) {
    const deck = new Deck();
    const cardsPerPlayer = 9;
    this.unplayed = deck.cards.slice(cardsPerPlayer * players.length);

    this.players = players.map((player, i) => {
      const startIndex = cardsPerPlayer * i;
      const endIndex = startIndex + cardsPerPlayer;
      return new Player({
        name: player.name,
        cards: deck.cards.slice(startIndex, endIndex)
      });
    });
  }
}
