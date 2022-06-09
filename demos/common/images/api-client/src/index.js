var express = require('express');
var bodyParser = require('body-parser');
var net = require('net'); 

var app = express();
var urlencodedParser = bodyParser.urlencoded({ extended: false })

var client = new net.Socket(); 
const port = 8001;
const host = '10.104.54.11';

const sendData = (action, primitive, args) => {
    client.connect({ port: port, host: host });
    client.write(`{"Action": "${action}", "Target": "${primitive}", "Arguments": [${args}]}`);
    client.write("\n");
    client.on('data', function(resp) {
        console.log(`Data received from the server: ${resp.toString()}.`);
        client.end();
    });
}

app.set('view engine', 'ejs');

app.get('/', function (req, res) {
    res.render("index")
});

app.get('/command', function (req, res) {
    res.render("command")
});

app.post('/command', urlencodedParser, function (req, res) {
    result = sendData(req.body.action, req.body.primitive, req.body.arguments);
    res.render("command");
});

app.use(express.static(__dirname + '/public'));

app.listen(1337);