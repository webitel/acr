var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter'),
    dilplan = require('./middleware/dialplan'),
    globalCollection = require('./middleware/system'),
    DEFAULT_HANGUP_CAUSE = require('./const').DEFAULT_HANGUP_CAUSE,
    Consul = require('consul');
    call = 0,
    internalExtension = require('./middleware/dialplan/internalExtansion');

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

    log.trace('New call %s', id);
    //console.log(conn.channelData.serialize());

    try {
        var context = conn.channelData.getHeader('Channel-Context'),
            destinationNumber = conn.channelData.getHeader('Channel-Destination-Number') ||
                conn.channelData.getHeader('Caller-Destination-Number');

        if (context == PUBLIC_CONTEXT) {
            dilplan.findActualPublicDialplan(destinationNumber, function (err, result) {
                if (err) {
                    // TODO
                    log.error(err.message);
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return
                };

                if (result.length == 0) {
                    log.error("Not found route");
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return;
                };

                globalCollection.getGlobalVariables(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        // TODO
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    };

                    if (result.length == 0) {
                        // TODO
                        log.error('Not found the route.');
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return;
                    };

                    conn.execute('set', 'domain_name=' + result[0]['domain']);
                    conn.execute('set', 'presence_data=' + result[0]['domain']);

                    var callflow = result[0]['callflow'];
                    var _router = new CallRouter(conn, globalVariable, result[0]['destination_number'], destinationNumber,
                        result[0]['timezone'], result[0]['version']);
                    try {
                        _router.start(callflow);
                    } catch (e) {
                        log.error(e.message);
                        //TODO узнать что ответить на ошибку
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    };

                });
            });
        } else {
            var domainName = conn.channelData.getHeader('variable_domain_name'),
                _isNotRout = true;
            dilplan.findActualDefaultDialplan(domainName, function (err, result) {

                if (err) {
                    log.error(err.message);
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return;
                };

                if (result.length == 0) {
                    log.warn("Not found route");
                };
                globalCollection.getGlobalVariables(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        // TODO
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    };

                    if (result instanceof Array) {
                        var _r, _reg;
                        for (var i = 0, len = result.length; i < len; i++) {
                            if (result[i]['destination_number'] && typeof result[i]['destination_number'] === 'string') {
                                _r = result[i]['destination_number'].match(new RegExp('^/(.*?)/([gimy]*)$'));
                                // Bad destination reg exp value
                                if (!_r) {
                                    _r = [null, result[i]['destination_number']]
                                };

                                _reg = new RegExp(_r[1], _r[2]).exec(destinationNumber);
                                if (_reg) {
                                    var callflow = result[i]['callflow'];
                                    var _router = new CallRouter(conn, globalVariable, result[i]['destination_number'],
                                        destinationNumber, result[i]['timezone'], result[0]['version']);
                                    try {
                                        _isNotRout = false;
                                        _router.start(callflow);
                                        break;
                                    } catch (e) {
                                        log.error(e.message);
                                        // TODO узнать что ответить на ошибку
                                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                                        break;
                                    };
                                };
                                log.trace('Break: %s', result[i]['destination_number']);
                            } else {
                                log.warn('Bad destination_number parameters');
                            };
                        };
                    };
                    if (_isNotRout) {
                        internalExtension(conn, destinationNumber, domainName);
                    };
                });
            });
        };
    } catch (e) {
        log.error(e.message);
        // TODO узнать что ответить на ошибку
        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
    };

    conn.on('esl::end', function() {
        log.trace("Call end %s", id);
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});