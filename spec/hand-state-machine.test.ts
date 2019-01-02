import { HandState } from "./../src/types";
import { getNextHandState } from "./../src/hand-state-machine";

describe("Hand state machine", () => {
  test("initial state", () => {
    const next = getNextHandState(null);
    expect(next).toEqual(HandState.A);
  });

  describe("Given there are cards remaining", () => {
    describe("state A", () => {
      test("moves to state B when hand count > 3", () => {
        const next = getNextHandState(HandState.A, {
          nextHandCount: 4,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.B);
      });

      test("remains in state A when hand count < 3", () => {
        const next = getNextHandState(HandState.A, {
          nextHandCount: Math.floor(Math.random() * 2) + 1,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("remains in state A when hand count === 3", () => {
        const next = getNextHandState(HandState.A, {
          nextHandCount: 3,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("remains in state A when hand count === 0", () => {
        const next = getNextHandState(HandState.A, {
          nextHandCount: 0,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("throws an error if inputs are invalid", () => {
        const fn = () =>
          getNextHandState(HandState.A, {
            nextHandCount: -1,
            cardsRemaining: true
          });

        expect(fn).toThrowError("Invalid hand count");
      });
    });

    describe("state B", () => {
      test("remains in state B when hand count > 3", () => {
        const next = getNextHandState(HandState.B, {
          nextHandCount: 4,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.B);
      });

      test("move to state A when hand count === 3", () => {
        const next = getNextHandState(HandState.B, {
          nextHandCount: 3,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("move to state A when hand count < 3", () => {
        const next = getNextHandState(HandState.B, {
          nextHandCount: Math.floor(Math.random() * 2) + 1,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("move to state A when hand count === 0", () => {
        const next = getNextHandState(HandState.B, {
          nextHandCount: 0,
          cardsRemaining: true
        });

        expect(next).toEqual(HandState.A);
      });

      test("throws an error if inputs are invalid", () => {
        const fn = () =>
          getNextHandState(HandState.B, {
            nextHandCount: -1,
            cardsRemaining: true
          });

        expect(fn).toThrowError("Invalid hand count");
      });
    });
  });

  describe("Given there are no cards remaining", () => {
    describe("state A", () => {
      const currentState = HandState.A;

      test("moves to state C if hand count > 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: Math.floor(Math.random() * 2) + 1,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.C);
      });

      test("moves to state D if hand count === 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: 0,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.D);
      });
    });

    describe("state B", () => {
      const currentState = HandState.B;

      test("moves to state C if hand count > 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: 4,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.C);
      });

      test("moves to state D if hand count === 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: 0,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.D);
      });

      test("throws an error if inputs are invalid", () => {
        const fn = () =>
          getNextHandState(currentState, {
            nextHandCount: -1,
            cardsRemaining: false
          });

        expect(fn).toThrowError("Invalid hand count");
      });
    });

    describe("state C", () => {
      const currentState = HandState.C;

      test("remains in state C if hand count > 0", () => {
        let next = getNextHandState(currentState, {
          nextHandCount: 4,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.C);

        next = getNextHandState(currentState, {
          nextHandCount: 3,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.C);
      });

      test("moves to state D if hand count === 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: 0,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.D);
      });

      test("throws an error if inputs are invalid", () => {
        let fn = () =>
          getNextHandState(currentState, {
            nextHandCount: -1,
            cardsRemaining: false
          });

        expect(fn).toThrowError("Invalid hand count");

        fn = () =>
          getNextHandState(currentState, {
            nextHandCount: 0,
            cardsRemaining: true
          });

        expect(fn).toThrowError("Invalid value for `cardsRemaining`");
      });
    });

    describe("state D", () => {
      const currentState = HandState.D

      test("moves to state C if hand count > 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: Math.floor(Math.random() * 10) + 1,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.C);
      })

      test("remains in state D if hand count === 0", () => {
        const next = getNextHandState(currentState, {
          nextHandCount: 0,
          cardsRemaining: false
        });

        expect(next).toEqual(HandState.D);
      })

      test("throws an error if inputs are invalid", () => {
        let fn = () =>
          getNextHandState(currentState, {
            nextHandCount: -1,
            cardsRemaining: false
          });

        expect(fn).toThrowError("Invalid hand count");

        fn = () =>
          getNextHandState(currentState, {
            nextHandCount: 0,
            cardsRemaining: true
          });

        expect(fn).toThrowError("Invalid value for `cardsRemaining`");
      });
    })
  });
});
