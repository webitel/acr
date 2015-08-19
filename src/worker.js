var esl = require('modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    publicContext = require('./middleware/publicContext'),
    defaultContext = require('./middleware/defaultContext'),
    globalCollection = require('./middleware/system'),
    DEFAULT_HANGUP_CAUSE = require('./const').DEFAULT_HANGUP_CAUSE
    ;

var PUBLIC_CONTEXT = 'public';

var esl_server = new esl.Server({host: conf.get('server:host'), port: process.env['WORKER_PORT'] || 10030,
        myevents: false }, function() {
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.warn('Call %s error: %s', id, error.message);
    });
    log.trace('New call %s', id);
    //console.log(conn.channelData.serialize());
    try {
        var context = conn.channelData.getHeader('Channel-Context'),
            destinationNumber = conn.channelData.getHeader('Channel-Destination-Number') ||
                conn.channelData.getHeader('Caller-Destination-Number');
        log.debug('Call %s -> %s', id, destinationNumber);

        globalCollection.getGlobalVariables(conn, conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
            if (err) {
                log.error(err.message);
                conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                return
            };

            var soundPref = '\/$${sounds_dir}\/en\/us\/callie';
            if (conn.channelData.getHeader('variable_default_language') == 'ru') {
                soundPref = '\/$${sounds_dir}\/ru\/RU\/elena';
            };

            conn.execute('set', 'sound_prefix=' + soundPref);

            if (context == PUBLIC_CONTEXT) {
                publicContext(conn, destinationNumber, globalVariable, !conn.channelData.getHeader('variable_webitel_direction'));
            } else {
                defaultContext(conn, destinationNumber, globalVariable, !conn.channelData.getHeader('variable_webitel_direction'));
            };

        });

    } catch (e) {
        log.error(e.message);
        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
    };

    conn.on('esl::end', function() {
        log.trace("Call end %s", id);
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});

process.on('uncaughtException', function (err) {
    log.error('UncaughtException:', err.message);
    log.error(err.stack);
    process.exit(1);
});