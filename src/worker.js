var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter');

var data = {
    "destination_number": "111",
    "domain": "",
    "context": "",
    "extension":
        [
            {
                "name": "e1",
                "if": // ::condition
                {
                    "expression": "2 == 1 || 1 != 2",
                    "then": [
                        {
                            "app": "answer"
                        },
                        {
                            "app": "echo"
                        },
                        {
                            "app": "info",
                            "data": "HELLO NODE ROUTER"
                        },
                        {
                            "if": {
                                "expression": "2==2",
                                "then": [
                                    {
                                        "app": "set",
                                        "data": "hello node app"
                                    }
                                ]
                            }
                        }
                    ],
                    "else": []
                }
            }
        ]
};

var esl_server = new esl.Server({host: conf.get('server:host'), port: process.env['WORKER_PORT'] || 10025, myevents:false}, function(){
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.error(error);
    });

    var extension = data['extension'];
    var _router = new CallRouter(conn);
    _router.doExec(extension);

/*
    log.trace('new call ' + id);
    conn.call_start = new Date().getTime();
    //log.log(conn.channelData.serialize('plain'));
    conn.execute('log', "hello node pid: " + process.pid);
    conn.execute('answer');
    conn.execute('echo', function(){
        log.trace('echoing');
    });
*/
    conn.on('esl::end', function(evt, body) {
        //this.call_end = new Date().getTime();
        //var delta = (this.call_end - this.call_start) / 1000;
        //log.trace("Call duration " + delta + " seconds");
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});