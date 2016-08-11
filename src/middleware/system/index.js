/**
 * Created by i.navrotskyj on 30.01.2015.
 */

var log = require('../../lib/log')(module),
    conf = require('../../conf'),
    globalVariables = {};

var sys = {
    getGlobalVariables: function (conn, uuid, cb) {

        if (globalVariables[uuid]) {
            cb(null, globalVariables[uuid]);
            return;
        }

        conn.api('global_getvar', function (globalVarObject) {
            try {
                var _json = {},
                    _param;
                var _body = globalVarObject['body'];
                if (_body) {
                    _body.split('\n').forEach(function (str) {
                        _param = str.split('=');
                        if (_param[0] == '') return;
                        _json[_param[0]] = _param[1];
                    });
                }
                
                globalVariables[uuid] = _json;
                log.info('Add hash global variable. Core uuid: ' + uuid);
                cb(null, globalVariables[uuid]);
            } catch (e) {
                log.error(e['message']);
                cb(e);
            }
        });
    }
};

module.exports = sys;