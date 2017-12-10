export interface Card {
  value: number;
  suit: string;
  name: string;
  visibleToAll: boolean;
  inPlay: boolean;
}

export enum Suit {
  Hearts = "Hearts",
  Clubs = "Clubs",
  Diamonds = "Diamonds",
  Spades = "Spades"
}
