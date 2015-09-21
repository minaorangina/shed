var deck = require('./initDeck');

function displayCards(card) {
    var img = document.createElement("img");
    img.src = card.image;
    div.appendChild(img);
}

var div = document.getElementsByClassName("pic")[0];
deck().forEach(displayCards);
