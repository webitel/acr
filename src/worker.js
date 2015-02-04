var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter'),
    dilplan = require('./middleware/dialplan'),
    globalCollection = require('./middleware/system'),
    call = 0;

var INBOUND_CONTEXT = 'default';

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

    if (context == INBOUND_CONTEXT) {
        dilplan.findActualPublicDialplan(conn.channelData.getHeader('Channel-Destination-Number'), function (err, result) {
            if (err) {
                // TODO
                log.error(err.message);
                conn.execute('hangup', 'DESTINATION_OUT_OF_ORDER');
                return
            };
            globalCollection.getGlobalVarFromUUID(conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                if (err) {
                    // TODO
                    log.error(err.message);
                    conn.execute('hangup', 'DESTINATION_OUT_OF_ORDER');
                    return
                };
                console.log('New Call ' + (call++));
                if (result.length == 0) {
                    // TODO
                    log.error('Error: Not found the route.');
                    conn.execute('hangup', 'DESTINATION_OUT_OF_ORDER');
                    return
                };

                var callflow = result[0]['callflow'];
                var _router = new CallRouter(conn, globalVariable);
                _router.start(callflow);

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