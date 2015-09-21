var Hapi = require('hapi');
var server = new Hapi.Server();
server.connection({
    port: process.env.PORT || 3000,
    host: 'localhost'
});

server.route(require('./routes'));

server.start(function() {
    console.log("Game server running at port ", server.info.uri);
});
