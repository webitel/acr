/**
 * Created by igor on 14.05.16.
 */

'use strict';


var log = require('./../lib/log')(module),
    dialplan = require('./dialplan'),
    DEFAULT_HANGUP_CAUSE = require('../const').DEFAULT_HANGUP_CAUSE,
    CallRouter = require('./callRouter');

module.exports = function (conn, destinationNumber, globalVariable, notExistsDirection) {
    let domainName = conn.channelData.getHeader('variable_domain_name');

    dialplan.findDialerDialplan(destinationNumber, domainName, (err, res) => {
        if (err) {
            log.error(err.message);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        }

        if (!res || !(res._cf instanceof Array)) {
            log.error(`Not found dialer ${destinationNumber} context`);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        }

        let callflow = res._cf;
        var _router = new CallRouter(conn, {
            "globalVar": globalVariable,
            "desNumber": destinationNumber,
            "chnNumber": destinationNumber,
            "timeOffset": null,
            "versionSchema": 2,
            "domain": domainName
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
