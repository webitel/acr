/**
 * Created by i.navrotskyj on 28.04.2015.
 */

let log = require('../lib/log')(module),
    dialplan = require('./dialplan'),
    CallRouter = require('./callRouter'),
    conf = require('../conf'),
    DEFAULT_PUBLIC_ROUTE = conf.get('defaultPublicRout'),
    DEFAULT_HANGUP_CAUSE = require('../const').DEFAULT_HANGUP_CAUSE
    ;

module.exports = function (conn, destinationNumber, globalVariable, notExistsDirection) {

    findNumber(destinationNumber, (e, result) => {
        if (e)
            return conn.execute('hangup', DEFAULT_HANGUP_CAUSE);

        if (result)
            return exec(result);

        const defaultPublicRout = DEFAULT_PUBLIC_ROUTE || globalVariable.webitel_default_public_route;

        if (!result && !defaultPublicRout) {
            log.warn("Not found route PUBLIC");
            return conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
        }

        return findNumber(defaultPublicRout, (e, result) => {
            if (e)
                return conn.execute('hangup', DEFAULT_HANGUP_CAUSE);

            if (result)
                return exec(result);


            log.warn(`Not found route default PUBLIC: ${defaultPublicRout}`);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
        });
    });


    const exec = (result) => {
        // WTEL-183
        if (notExistsDirection) {
            log.trace('set: webitel_direction=inbound');
            conn.execute('set', 'webitel_direction=inbound');
        }

        if (result[0]['fs_timezone']) {
            conn.execute('set', 'timezone=' + result[0]['fs_timezone']);
        }

        conn.execute('set', 'domain_name=' + result[0]['domain']);
        conn.execute('set', 'force_transfer_context=default');
        conn.execute('hash', 'insert/' + result[0]['domain']+'-last_dial/global/${uuid}');

        let callflow = result[0]['callflow'];
        let _router = new CallRouter(conn, {
            "globalVar": globalVariable,
            "desNumber": result[0]['destination_number'],
            "chnNumber": destinationNumber,
            "timeOffset": result[0]['fs_timezone'],
            "versionSchema": result[0]['version'],
            "domain": result[0]['domain'],
            "onDisconnect": result[0]['onDisconnect']
        });

        try {
            log.trace('Exec: %s', destinationNumber);
            _router.run(callflow);
        } catch (e) {
            log.error(e.message);
            //TODO узнать что ответить на ошибку
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
        }
    };
};

const findNumber = (number, cb) => {
    dialplan.findActualPublicDialplan(number, (err, result) => {
        if (err) {
            log.error(e.message);
            return cb(err);
        }

        return cb(null, result.length === 0 ? null : result)
    });
};