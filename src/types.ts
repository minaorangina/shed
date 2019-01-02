import { type } from "os";

export interface Card {
  value: number;
  suit: string;
  name: string;
  visibleToAll?: boolean;
  inPlay?: boolean;
}

export enum HandState {
  A = "3 cards",
  B = ">3 cards",
  C = ">0 cards",
  D = "0 cards"
}

export interface State {
  current: ValidStates;
}

export interface HandInputs {
  cardsRemaining: boolean,
  nextHandCount: number
}

export type NextState = ValidStates; // union type of all state types

export type ValidStates = HandState;

export enum Suit {
  Hearts = "Hearts",
  Clubs = "Clubs",
  Diamonds = "Diamonds",
  Spades = "Spades"
}
