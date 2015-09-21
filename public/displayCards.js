

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
