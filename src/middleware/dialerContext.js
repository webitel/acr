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

        conn.subscribe('CHANNEL_ANSWER');
        // TODO Channel not answer
//originate {origination_uuid=b55f4a2f-7953-4964-b536-79203314a2e3,dlr_queue=572a170e576151df0d6b164a,domain_name=10.10.10.144,origination_caller_id_number=1234567,origination_caller_id_name=Igor2,gatewayPositionMap=0>0}sofia/gateway/test/380730367300 &socket(10.10.10.25:10030 async full)
        conn.on('esl::event::CHANNEL_ANSWER::*', (res) => {
            console.log(res);
            log.trace('onAnswer');
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
    });
};
