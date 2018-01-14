import Player from "../src/player";
import { orderedCards } from "./fixtures";

describe("Player", () => {
  const penny = new Player({ name: "Penny" });

  describe(".startingHand", () => {
    test("Initialises player's cards", () => {
      penny.startingHand = [...orderedCards].slice(0, 9);
      expect(penny.handCards.length).toBe(3);
      expect(penny.tableCards.length).toBe(6);
    });
  });

  describe(".cards", () => {
    test("Returns player's table cards and hand cards", () => {
      expect(penny.cards).toHaveProperty("tableCards");
      expect(penny.cards).toHaveProperty("handCards");
    });
  });
});
