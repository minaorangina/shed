import GameEngine from "../src/game-engine";
import Player from "../src/player";
import { orderedCards } from "./fixtures";

describe("Game Engine", () => {
  const playerNames = ["Nancy", "Mae"];
  const engine = new GameEngine(playerNames);

  describe("Initial state", () => {
    test("Requires min 2 players", () => {
      function makeGameEngine() {
        return new GameEngine([]);
      }
      expect(makeGameEngine).toThrow();
    });

    test("Correct number of unplayed cards", () => {
      const expected = 52 - playerNames.length * 9;
      expect(engine.unplayed.length).toBe(expected);
    });

    test("Players dealt correct number of cards", () => {
      const cardsInPlay = engine.getCardsInPlay();
      cardsInPlay.forEach(playerCards => {
        expect(playerCards.hand.length).toEqual(3);
        expect(playerCards.tableCards.length).toEqual(6);
      });
    });

    test("Players' hand cards are visible", () => {
      const cardsInPlay = engine.getCardsInPlay();
      cardsInPlay.forEach(playerCards => {
        playerCards.hand.forEach(card => {
          expect(card.visibleToAll).toBeTruthy()
        })
      });
    });
  });
  test("dummy test", () => {
    expect(1 + 1).toEqual(2);
  });
});
