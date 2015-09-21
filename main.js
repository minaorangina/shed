(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){


function displayCards(cardArray) {
    console.log(cardArray);
    cardArray.forEach(function(card) {
        var img = document.createElement("img");
        img.src = card.image;
        div.appendChild(img);
    });

}

var div = document.getElementsByClassName("pic")[0];

module.exports = displayCards;

},{}],2:[function(require,module,exports){
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

},{"./shuffle":5}],3:[function(require,module,exports){

},{}],4:[function(require,module,exports){
arguments[4][3][0].apply(exports,arguments)
},{"dup":3}],5:[function(require,module,exports){
function shuffle(array) {
  var currentIndex = array.length,
      temporaryValue,
      randomIndex ;

    // While there remain elements to shuffle...
    while (0 !== currentIndex) {

        // Pick a remaining element...
        randomIndex = Math.floor(Math.random() * currentIndex);
        currentIndex -= 1;

        // And swap it with the current element.
        temporaryValue = array[currentIndex];
        array[currentIndex] = array[randomIndex];
        array[randomIndex] = temporaryValue;
    }
    return array;
}
module.exports = shuffle;

},{}],6:[function(require,module,exports){
var deck = require('./initDeck');
var displayCards = require('./displayCards');

displayCards(deck.slice(0,3));

},{"./displayCards":1,"./initDeck":2}]},{},[1,2,3,4,5,6]);
