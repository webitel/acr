/**
 * Created by Igor Navrotskyj on 20.08.2015.
 */

var log = require('../lib/log')(module);

'use strict';

module.exports = function (CallRouter) {

    CallRouter.prototype.__disa = function (app, cb) {
        try {
            var scope = this;
            var prop = app['disa'],
                callerID = '',
                channelCallerIdNumber = scope.getChnVar('Channel-Caller-ID-Number'),
                password = '',
                isCallBack = prop['callback'] || false,
                destinationNumber = '',
                isAuth = prop['auth'],
                timeout= prop['timeout'] || 5000
                ;

            scope.__answer({
                "answer": ""
            });

            var call = function () {
                setDestinationNumber(function () {
                    try {
                        if (destinationNumber == '') {
                            log.warn('Bar destination number.');
                            return cb && cb();
                        };
                        if (isCallBack) {
                            var gateway = prop['gateway'];
                            if (gateway) {
                                var gwMask = gateway['mask'],
                                    gwName = gateway['name'],
                                    gwDialString = gateway['dialString'] || '',
                                    dialString = '',
                                    _reg
                                    ;
                                if (gwMask) {
                                    var _r = gwMask.match(new RegExp('^/(.*?)/([gimy]*)$'));
                                    // Bad destination reg exp value
                                    if (!_r) {
                                        _r = [null, gwMask]
                                    };
                                    try {
                                        _reg = new RegExp(_r[1], _r[2]).exec(channelCallerIdNumber);
                                    } catch (e) {
                                        _reg = null;
                                    };
                                };

                                if (_reg) {
                                    gwDialString = gwDialString.replace(/\$(\d+)/g, function (a) {
                                        return _reg[parseInt(a.substring(1))] || '';
                                    });
                                };

                                dialString = 'sofia/gateway/' + gwName + '/' + gwDialString;

                            } else {
                                dialString = 'user/' + callerID;
                            };

                            dialString = '[origination_caller_id_number=' + destinationNumber + ']' + dialString;

                            scope.__hangup({
                                "hangup": "NORMAL_CLEARING"
                            });

                            var api = ''.concat("sched_api +1 none originate '", dialString, "' '", callerID != '' ? 'set_user:' + callerID + ',' : '',
                                "transfer:", destinationNumber, " XML default' inline ");

                            log.trace(api);

                            scope.connection.api(api);

                        } else {
                            scope.__goto({
                                "goto": "default:" + destinationNumber
                            });
                        };
                        return cb && cb();
                    } catch (e) {
                        log.error(e);
                    }
                });
            };

            var setDestinationNumber = function (cbDestinationNumber) {
                scope.__playback({
                    "playback": {
                        "name": "ivr/ivr-please_enter_extension_followed_by_pound.wav",
                        "type": "local",
                        "getDigits": {
                            "setVar": "webitel_disa_destination",
                            "min": 1,
                            "max": 16,
                            "timeout": timeout
                        }
                    }
                }, function () {
                    destinationNumber = scope.getChnVar('webitel_disa_destination');
                    cbDestinationNumber();
                });
            };

            var playErrorAuth = function () {
                scope.__playback({
                    "playback": {
                        "name": "voicemail/vm-fail_auth.wav",
                        "type": "local"
                    }
                });
            };

            if (isAuth) {
                scope.__playback({
                    "playback": {
                        "name": "voicemail/vm-enter_id.wav",
                        "type": "local",
                        "getDigits": {
                            "setVar": "webitel_disa_user",
                            "min": 1,
                            "max": 10,
                            "timeout": timeout
                        }
                    }
                }, function () {
                    callerID = scope.getChnVar('webitel_disa_user');
                    if (callerID == '') {
                        // TODO playback bad user
                        log.warn('Bad callerId.');
                        return cb && cb();
                    }
                    ;
                    callerID += '@' + scope.domain;
                    scope
                        .connection
                        .api('user_data ' + callerID + ' variable vm-password', function (res) {
                            password = res['body'];
                            if (password == '' || password.indexOf('-ERR') == 0) {
                                playErrorAuth();
                                log.warn('Bad user parameters password.');
                                return cb && cb();
                            }
                            ;

                            scope.__playback({
                                "playback": {
                                    "name": "voicemail/vm-enter_pass.wav",
                                    "type": "local",
                                    "getDigits": {
                                        "setVar": "webitel_disa_password",
                                        "min": 1,
                                        "max": 10,
                                        "timeout": timeout
                                    }
                                }
                            }, function () {
                                if (password !== scope.getChnVar('webitel_disa_password')) {
                                    playErrorAuth();
                                    log.warn('Bad password.');
                                    return cb && cb();
                                }
                                ;
                                call();
                            })
                        }
                    );
                });
            } else {
                call();
            }
            ;
        } catch (e) {
            log.error(e);
            return cb && cb();
        }
    };
};