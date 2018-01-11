import GameEngine from "../src/game-engine";

describe("Game Engine", () => {
  const players = [{ name: "Nancy" }, { name: "Mae" }];
  const engine = new GameEngine(players);

  describe("Initial state", () => {
    test("Requires min 2 players", () => {
      function makeGameEngine() {
        return new GameEngine([]);
      }
      expect(makeGameEngine).toThrow();
    });

    test("Correct number of unplayed cards", () => {
      const expected = 52 - players.length * 9;
      expect(engine.game.unplayed.length).toBe(expected);
    });

    test("Players dealt correct number of cards", () => {
      expect(engine.game.players.length).toBe(players.length);
      engine.game.players.forEach(player => {
        expect(player.tableCards.length).toBe(6);
      });
      engine.game.players.forEach(player => {
        expect(player.handCards.length).toBe(3);
      });
    });
  });
});
