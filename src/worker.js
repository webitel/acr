var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter'),
    dilplan = require('./middleware/dialplan'),
    globalCollection = require('./middleware/system'),
    DEFAULT_HANGUP_CAUSE = require('./const').DEFAULT_HANGUP_CAUSE,
    Consul = require('consul');
    call = 0;

var PUBLIC_CONTEXT = 'public';

/*
var consul = new Consul({
    host: "10.10.10.160"
});

var check = {
    name: 'ACR',
    ttl: '15s',
    notes: 'Started'
};

consul.agent.check.register(check, function(err) {
    if (err) throw err;
    consul.agent.check.pass('ACR', function(err) {
        if (err) throw err;
        return {
            Output: "test"
        }
    });
    setInterval(function () {
        consul.agent.check.pass('ACR', function(err) {
            if (err) throw err;
            return {
                Output: "test"
            }
        });
    }, 15000)

});
*/


var esl_server = new esl.Server({host: conf.get('server:host'), port: process.env['WORKER_PORT'] || 10025,
        myevents: false }, function() {
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.error(error.message);
    });

    //console.log(conn.channelData.serialize());

    try {

        var context = conn.channelData.getHeader('Channel-Context'),
            destinationNumber = conn.channelData.getHeader('Channel-Destination-Number');
// TODO replace !==
        if (context == PUBLIC_CONTEXT) {
            dilplan.findActualPublicDialplan(destinationNumber, function (err, result) {
                if (err || result.length == 0) {
                    // TODO
                    var message = (err)
                        ? err.message
                        : "Not found route";
                    log.error(message);
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return
                }
                ;

                globalCollection.getGlobalVariables(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        // TODO
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    }
                    ;

                    if (result.length == 0) {
                        // TODO
                        log.error('Error: Not found the route.');
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    };
                    conn.execute('set', 'domain_name=' + result[0]['domain']);
                    var callflow = result[0]['callflow'];
                    var _router = new CallRouter(conn, globalVariable, result[0]['destination_number'], destinationNumber);
                    try {
                        _router.start(callflow);
                    } catch (e) {
                        log.error(e.message);
                        //TODO узнать что ответить на ошибку
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    }
                    ;

                });
            });
        } else {
            dilplan.findActualDefaultDialplan(conn.channelData.getHeader('variable_domain_name'), function (err, result) {
                if (err || result.length == 0) {
                    // TODO
                    var message = (err)
                        ? err.message
                        : "Not found route";
                    log.error(message);
                };
                globalCollection.getGlobalVariables(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        // TODO
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    }
                    ;

                    if (result instanceof Array) {
                        var _r, _reg;
                        for (var i = 0, len = result.length; i < len; i++) {
                            if (result[i]['destination_number']) {
                                _r = result[i]['destination_number'].match(new RegExp('^/(.*?)/([gimy]*)$'));
                                // Bad destination reg exp value
                                if (!_r) {
                                    _r = [null, result[i]['destination_number']]
                                };

                                _reg = new RegExp(_r[1], _r[2]).exec(destinationNumber);
                                if (_reg) {
                                    var callflow = result[i]['callflow'];
                                    var _router = new CallRouter(conn, globalVariable, result[i]['destination_number'], destinationNumber);
                                    try {
                                        _router.start(callflow);
                                        break;
                                    } catch (e) {
                                        log.error(e.message);
                                        // TODO узнать что ответить на ошибку
                                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                                        break;
                                    }
                                    ;
                                }
                                ;
                                log.trace('Break: %s', result[i]['destination_number']);
                            }
                            ;
                        }
                        ;
                    }
                    ;
                });
            });
        }
        ;
    } catch (e) {
        log.error(e.message);
        // TODO узнать что ответить на ошибку
        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
    };

    conn.on('esl::end', function(evt, body) {
        log.trace("Call end");
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});