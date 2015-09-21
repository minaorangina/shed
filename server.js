var Hapi = require('hapi');
var server = new Hapi.server();
server.connection({
    port: process.env.PORT || 3000
});

server.start(function() {
    console.log("Game server running at port ", server.info.uri);
});
