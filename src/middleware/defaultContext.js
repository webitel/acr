/**
 * Created by i.navrotskyj on 28.04.2015.
 */

var log = require('../lib/log')(module),
    dialplan = require('./dialplan'),
    CallRouter = require('./callRouter'),
    DEFAULT_HANGUP_CAUSE = require('../const').DEFAULT_HANGUP_CAUSE,
    internalExtension = require('./dialplan/internalExtansion')
    ;

function setupPickupParameters(conn, userId, domain) {
    if (!userId || !domain || !conn) {
        return log.error('Bad parameters setupPickupParameters');
    };

    conn.execute('export', 'dialed_extension=' + userId);
    conn.execute('hash', 'insert/' + domain + '-call_return/' + userId + '/${caller_id_number}');
    conn.execute('hash', 'insert/' + domain + '-last_dial_ext/' + userId + '/${uuid}');
    conn.execute('hash', 'insert/' + domain + '-last_dial_ext/global/${uuid}');
}

module.exports = function (conn, destinationNumber, globalVariable, notExistsDirection) {
    var domainName = conn.channelData.getHeader('variable_domain_name'),
        _isNotRout = true,
        _callerIdNumber = conn.channelData.getHeader('Channel-Caller-ID-Number')
        ;
    conn.execute('unset', 'sip_h_call-info');

    if (_callerIdNumber)
        conn.execute('hash', 'insert/spymap/${domain_name}-' + _callerIdNumber + '/${uuid}');

    dialplan.findActualExtension(destinationNumber, domainName, function (err, resultExtension) {
        if (err) {
            log.error(err.message);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return;
        }

        if (resultExtension) {
            try {
                // WTEL-183
                if (notExistsDirection) {
                    var _tmpDirection = conn.channelData.getHeader('variable_user_scheme')
                        ? 'internal'
                        : 'outbound';
                    conn.execute('set', 'webitel_direction=' + _tmpDirection);
                    log.trace('set: webitel_direction=%s', _tmpDirection);

                }
                if (resultExtension['fs_timezone']) {
                    conn.execute('set', 'timezone=' + resultExtension['fs_timezone']);
                }

                setupPickupParameters(conn, resultExtension['destination_number'], resultExtension['domain']);

                var callflow = resultExtension['callflow'];
                var _router = new CallRouter(conn, {
                    "globalVar": globalVariable,
                    "desNumber": resultExtension['destination_number'],
                    "chnNumber": destinationNumber,
                    "timeOffset": resultExtension['fs_timezone'],
                    "versionSchema": resultExtension['version'],
                    "domain": resultExtension['domain'],
                    "onDisconnect": resultExtension['onDisconnect']
                });
                _router.run(callflow);
            } catch (e) {
                log.error(e.message);
                conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            }
        } else {
            dialplan.findActualDefaultDialplan(domainName, function (err, result) {

                if (err) {
                    log.error(err.message);
                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                    return;
                }

                if (result.length == 0) {
                    log.warn("Not found route DEFAULT");
                }

                // WTEL-183
                if (notExistsDirection) {
                    conn.execute('set', 'webitel_direction=outbound');
                    log.trace('set: webitel_direction=outbound');
                }

                if (result instanceof Array) {
                    var _r, _reg;
                    for (var i = 0, len = result.length; i < len; i++) {
                        if (result[i]['destination_number'] && typeof result[i]['destination_number'] === 'string') {
                            _r = result[i]['destination_number'].match(new RegExp('^/(.*?)/([gimy]*)$'));
                            // Bad destination reg exp value
                            if (!_r) {
                                _r = [null, result[i]['destination_number']]
                            }
                            try {
                                _reg = new RegExp(_r[1], _r[2]).exec(destinationNumber);
                            } catch (e) {
                                _reg = null;
                            }
                            if (_reg) {
                                log.trace('Exec: %s', result[i]['destination_number']);

                                if (result[i]['fs_timezone']) {
                                    conn.execute('set', 'timezone=' + result[i]['fs_timezone']);
                                }

                                var callflow = result[i]['callflow'];
                                var _router = new CallRouter(conn, {
                                    "globalVar": globalVariable,
                                    "desNumber": result[i]['destination_number'],
                                    "chnNumber": destinationNumber,
                                    "timeOffset": result[i]['fs_timezone'],
                                    "versionSchema": result[i]['version'],
                                    "domain": result[i]['domain'],
                                    "onDisconnect": result[i]['onDisconnect']
                                });

                                try {
                                    _isNotRout = false;
                                    _router.run(callflow);
                                    break;
                                } catch (e) {
                                    log.error(e.message);
                                    // TODO узнать что ответить на ошибку
                                    conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                                    break;
                                }
                            }
                            log.trace('Break: %s', result[i]['destination_number']);
                        } else {
                            log.warn('Bad destination_number parameters');
                        }
                    }
                }
                if (_isNotRout) {
                    setupPickupParameters(conn, destinationNumber, domainName);
                    internalExtension(conn, destinationNumber, domainName);
                }

            });
        }
    });
};