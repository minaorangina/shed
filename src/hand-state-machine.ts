import { NextState, HandState, HandInputs } from "./types";

export function getNextHandState(
  state: HandState,
  inputs?: HandInputs
): NextState {
  if (!inputs) {
    return HandState.A;
  }
  const { cardsRemaining, nextHandCount } = inputs;
  if (nextHandCount < 0) {
    throw new RangeError("Invalid hand count")
  }

  if (cardsRemaining === true) {
    if (state === HandState.C || state === HandState.D) {
      throw new Error("Invalid value for `cardsRemaining`")
    }

    if (nextHandCount > 3) {
      return HandState.B;
    }

    if (nextHandCount <= 3) {
      return HandState.A;
    }
  }

  if (cardsRemaining === false) {
    if (nextHandCount > 0) {
      return HandState.C
    }

    if (nextHandCount === 0) {
      return HandState.D
    }
  }
}
