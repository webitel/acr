var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    CallRouter = require('./lib/callRouter'),
    dilplan = require('./middleware/dialplan');

var esl_server = new esl.Server({host: conf.get('server:host'), port: process.env['WORKER_PORT'] || 10025,
        myevents: false }, function() {
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.error(error.message);
    });

    dilplan.findActualDialplan(conn.channelData.getHeader('Channel-Destination-Number'), function (err, result) {
        if (err) {
            // TODO
            conn.execute('hangup', 'NO_ROUTE_TRANSIT_NET');
            return
        };

        if (result.length == 0) {
            // TODO
            conn.execute('hangup', 'NO_ROUTE_TRANSIT_NET');
            return
        };
        conn.execute("set", "domain_name=" + result[0]['domain']);
        var callflow = result[0]['callflow'];
        var _router = new CallRouter(conn);
        _router.doExec(callflow);
    });
    //console.log(conn.channelData.getHeader('Channel-Destination-Number'));
    conn.on('esl::end', function(evt, body) {
        log.trace("Call end");
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});