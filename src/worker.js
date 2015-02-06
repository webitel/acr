var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter'),
    dilplan = require('./middleware/dialplan'),
    globalCollection = require('./middleware/system'),
    DEFAULT_HANGUP_CAUSE = require('./const').DEFAULT_HANGUP_CAUSE,
    Consul = require('consul');
    call = 0;

var INBOUND_CONTEXT = 'default';

//var consul = new Consul({
//    host: "10.10.10.160"
//});
//
//consul.agent.service.register('ACR', function(err) {
//    if (err) throw err;
//    log.info('Consul : start');
//});

var esl_server = new esl.Server({host: conf.get('server:host'), port: process.env['WORKER_PORT'] || 10025,
        myevents: false }, function() {
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.error(error.message);
    });

    //console.log(conn.channelData.serialize());

    var context = conn.channelData.getHeader('Channel-Context');

        //if (context == INBOUND_CONTEXT) {
        if (true) {
            dilplan.findActualPublicDialplan(conn.channelData.getHeader('Channel-Destination-Number'), function (err, result) {
                if (err) {
                    // TODO
                    log.error(err.message);
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return
                };

                globalCollection.getGlobalVarFromUUID(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        // TODO
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    }
                    ;
                    console.log('New Call ' + (call++));
                    if (result.length == 0) {
                        // TODO
                        log.error('Error: Not found the route.');
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    };

                    var callflow = result[0]['callflow'];
                    var _router = new CallRouter(conn, globalVariable);
                    try {
                        _router.start(callflow);
                    } catch (e) {
                        log.error(e.message);
                        // TODO узнать что ответить на ошибку
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    }

                });
            });
        } else {
            log.info('OUTBOUND');
        };

    conn.on('esl::end', function(evt, body) {
        log.trace("Call end");
        console.log('End Call ' + (call--));
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});