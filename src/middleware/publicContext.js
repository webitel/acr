/**
 * Created by i.navrotskyj on 28.04.2015.
 */

var log = require('../lib/log')(module),
    dialplan = require('./dialplan'),
    CallRouter = require('./callRouter'),
    DEFAULT_HANGUP_CAUSE = require('../const').DEFAULT_HANGUP_CAUSE
    ;

module.exports = function (conn, destinationNumber, globalVariable) {
    dialplan.findActualPublicDialplan(destinationNumber, function (err, result) {
        if (err) {
            // TODO
            log.error(err.message);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        };

        if (result.length == 0) {
            log.warn("Not found route PUBLIC");
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return;
        };

        conn.execute('set', 'domain_name=' + result[0]['domain']);
        conn.execute('set', 'presence_data=' + result[0]['domain']);
        conn.execute('set', 'force_transfer_context=default');

        var callflow = result[0]['callflow'];
        var _router = new CallRouter(conn, {
            "globalVar": globalVariable,
            "desNumber": result[0]['destination_number'],
            "chnNumber": destinationNumber,
            "timeOffset": result[0]['timezone'],
            "versionSchema": result[0]['version'],
            "domain": result[0]['domain']
        });

        try {
            log.trace('Exec: %s', destinationNumber);
            _router.run(callflow);
        } catch (e) {
            log.error(e.message);
            //TODO узнать что ответить на ошибку
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
        };
    });
};