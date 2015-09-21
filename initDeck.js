var shuffle = require('./shuffle');

function Card(suitName, value, image){
    this.suit = suitName;
    this.value = value;
    this.image = image;
}
Card.prototype.toString = function(){
    return this.value + " of " + this.suit;
};

function initDeck(){
    var deck = [];
    var suits = ["hearts", "diamonds", "spades", "clubs"];
    var edgeCards = {
        1: "ace",
        11: "jack",
        12: "queen",
        13: "king"
    };
    suits.forEach(function(suit){
        for (var i = 0; i < 13; i++){
            var cardValue = edgeCards[i+1] || i+1;
            var pathToImg = "img/" + cardValue + "_of_" + suit +".svg";
            deck.push(new Card(suit, i+1, pathToImg));
        }
    });
    return shuffle(deck);
}

module.exports = initDeck();
