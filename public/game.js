var deck = require('./initDeck');
var displayCards = require('./displayCards');
var faceDownCards = deck.slice(0,3);
var handCards = deck.slice(0,3);
displayCards(faceDownCards);
displayCards(handCards);
