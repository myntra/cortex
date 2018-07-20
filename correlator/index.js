var express = require('express');
var app = express();
var     morgan = require('morgan'),
    winston = require('winston'),
    bodyParser = require('body-parser');
var config = require('./config.json');
var routes = require('./routes')
var eventHandler = require('./handlers/eventHandler')

var handlers = {
  eventHandler: eventHandler
};

app.use(morgan(config.log.level.morgan));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({
  extended: true
}));

winston.level = config.log.level.winston;

var server = require('http').createServer(app);
routes.routes(app, handlers);
var port =  config.port
server.listen(port|| process.env.PORT);
console.log("REST API server listening on port %d in %s mode", port, app.settings.env);
