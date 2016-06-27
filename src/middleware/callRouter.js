/**
 * Created by i.navrotskyj on 24.01.2015.
 */

'use strict';

var log = require('./../lib/log')(module),
    httpReq = require('./httpRequest'),
    sms = require('./sms'),
    dbRoute = require('./dbRoute'),
    async = require('async'),
    findDomainVariables = require('./dialplan').findDomainVariables,
    updateDomainVariables = require('./dialplan').updateDomainVariables,
    findExtension = require('./dialplan').findActualExtension,
    blackList = require('./blackList'),
    Event = require('modesl').Event,
    calendar = require('./calendar/index');

const VARIABLES_MAP = require('./variablesMap');
const WEBITEL_RECORD_FILE_NAME = 'webitel_record_file_name';

const MEDIA_TYPE = {
    WAV: 'wav',
    MP3: 'mp3',
    SILENCE: 'silence',
    LOCAL: 'local',
    SHOUT: 'shout',
    TONE: 'tone',
    SAY: 'say'
};

const OPERATION = {
    IF: "if",
    THEN: "then",
    ELSE: "else",
    SWITCH: "switch",
    APPLICATION: "app",
    DATA: "data",
    ASYNC: "async",

    ECHO: "echo",

    ANSWER: "answer",
    SET: "setVar",
    GOTO: "goto",
   /* GATEWAY: "gateway",
    DEVICE: "device", */
    RECORD_SESSION: "recordSession",
    RECORD_FILE: "recordFile",
    HANGUP: "hangup",
    SCRIPT: "script",
    LOG: "log",
    HTTP: "httpRequest",
    SLEEP: "sleep",

    CONFERENCE: "conference",

    SCHEDULE: "schedule",

    BRIDGE: "bridge",
    PLAYBACK: "playback",

    BREAK: "break",

    CALENDAR: "calendar",
    PARK: "park",
    QUEUE: "queue",
    CC_POSITION: "ccPosition",

    EXPORT_VARS: "exportVars",
    VOICEMAIL: "voicemail",

    IVR: "ivr",

    BIND_ACTION: "bindAction",
    CLEAR_ACTION: "clearAction",

    BIND_EXTENSION: "bindExtension",

    ATT_XFER: 'attXfer',

    UN_SET: 'unSet',

    SET_USER: 'setUser',
    CALL_FORWARD: 'checkCallForward',
    RECEIVE_FAX: 'receiveFax',

    TAGS: 'setArray',
    BLACK_LIST: 'blackList',

    PICKUP: 'pickup',

    DISA: 'disa',

    SEND_SMS: 'sendSms',

    LOCATION: "geoLocation",
    RINGBACK: "ringback",

    SET_SOUNDS: "setSounds",
    EVENT: "event",

    IN_BAND_DTMF: 'inBandDTMF',
    FLUSH_DTMF: 'flushDTMF',
    EMAIL: 'sendEmail',
    MATH: 'math',
    STRING: 'string',

    EAVESDROP: 'eavesdrop',
    SIP_REDIRECT: 'sipRedirect',
    AGENT: 'agent',
    AVMD: "avmd",
    TELEGRAM: "telegram"
};

const FS_COMMAND = {
    ANSWER: "answer",
    PRE_ANSWER: "pre_answer",
    RING_READY: "ring_ready",
    TRANSFER: "transfer",
    HANGUP: "hangup",

    SET: "set",
    MULTISET: "multiset",
    EXPORT: "export",

    RECORD_SESSION: "record_session",
    RECORD: "record",
    STOP_RECORD_SESSION: "stop_record_session",

    LUA: "lua",
    JS: "js",

    LOG: "log",

    ECHO: "echo",
    DELAY_ECHO: "delay_echo",

    SLEEP: "sleep",

    CONFERENCE: "conference",

    SCHEDULE_HANGUP: "sched_hangup",
    SCHEDULE_TRANSFER: "sched_transfer",

    BRIDGE: "bridge",

    PLAYBACK: "playback",
    BROADCAST: "uuid_broadcast",

    PLAY_AND_GET: "play_and_get_digits",
    PARK: "valet_park",
    CALLCENTER: "callcenter",
    VOICEMAIL: "voicemail",

    IVR: 'ivr',

    BIND_DIGIT_ACTION: 'bind_digit_action',
    CLEAR_DIGIT_ACTION: 'clear_digit_action',

    BIND_EXTENSION: 'bind_meta_app',

    ATT_XFER: 'att_xfer',

    UN_SET: 'unset',

    SET_USER: 'set_user',

    RX_FAX: 'rxfax',

    PUSH: 'push',

    PICKUP: 'pickup',
    EVENT: "event",
    START_DTMF: 'start_dtmf',
    STOP_DTMF: 'stop_dtmf',
    FLUSH_DTMF: 'flush_dtmf',

    HASH: 'hash',
    EAVESDROP: 'eavesdrop',
    USERSPY: 'userspy',

    DEFLECT: 'deflect',
    REDIRECT: 'redirect',
    AVMD_START: "avmd_start",
    AVMD_STOP: "avmd_stop",
};


const COMMANDS = {
    REGEXP: "&reg"
};

const MAX_CYCLE_COUNT = 20;

var CallRouter = module.exports = function (connection, option) {
    option = option || {};
    this.index = 0;
    this.cycleCount = 0;
    this.globalVar = option['globalVar'] || {};
    this.connection = connection;
    connection.__callRouter = this;
    this.regCollection = {};
    this.offset = option['timeOffset'];
    this.domain = option['domain'];
    this.domainVariables = {};
    this.end = false;
    this.channelDestinationNumber = option['chnNumber'];
    this.updateDomainVariable = false;
    this.uuid = connection.channelData.getHeader('variable_uuid');
    this._dumpArrayIndex = {};

    //this.localVar = option['localVariables'] || {};
    //this._dbId = option['id'] || '';
    //this.COLLECTION_NAME = option['collectionName'];
    //this.updateLocalVariable = false;

    this.versionSchema = option['versionSchema'];
    this.setDestinationNumber(option['desNumber'], option['chnNumber']);

    //this.xData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');

    this.__setVar({
        "setVar": "eavesdrop_group=" + this.domain
    });

    this.log = {
        info: (msg) => {
            log.info(this.uuid + ' => ' + (msg || ''));
        },
        error: (msg) => {
            log.error(this.uuid + ' => ' + (msg || ''));
        },
        warn: (msg) => {
            log.warn(this.uuid + ' => ' + (msg || ''));
        },
        debug: (msg) => {
            log.warn(this.uuid + ' => ' + (msg || ''));
        }
    };
};

require('./email')(CallRouter, OPERATION.EMAIL);
require('./disa')(CallRouter);
require('./location/number')(CallRouter, OPERATION.LOCATION);

function push(arr, e) {
    arr.push(e);
    return arr.length - 1;
};

function equalsRange (_curentDay, _tmp, maxVal) {
    var _min, _max;

    _tmp = _tmp.split('-');
    _min = parseInt(_tmp[0]);
    _max = _tmp[1]
        ? parseInt(_tmp[1])
        : maxVal;
    if (_min > _max) {
        _tmp = _max;
        _max = _min;
        _min = _tmp;
    };
    return (_curentDay >= _min && _curentDay <= _max);
};

/* TODO

 CallRouter.prototype.getLocalVariable = function (key) {
 return this.localVar[key];
 };

 CallRouter.prototype.setLocalVariable = function (key, value) {
 this.localVar[key] = value;
 this.updateLocalVariable = true;
 };

 CallRouter.prototype.updateLocalVariables = function () {
 try {
 if (this.updateLocalVariable) {
 dbRoute.setLocalVariables(this._dbId, this.localVar, this.COLLECTION_NAME, function (err, res) {
 console.log(arguments);
 });
 return true;
 }
 ;
 return false;
 } catch (e) {
 log.error(e['message']);
 };
 };

 */

CallRouter.prototype.saveDomainVariables = function (cb) {
    try {
        if (this.updateDomainVariable) {
            updateDomainVariables(this.domain, this.domainVariables, function (err, res) {
                if (err) {
                    log.error('Save domain variable error: %s', err['message']);
                    if(cb) cb(err);
                    return
                };

                log.trace('Saved domain variable %s', res);
            });
        };
    } catch (e) {
        log.error(e['message']);
    };
};

CallRouter.prototype.setupDomainVariables = function (cb) {
    var scope = this;
    findDomainVariables(this.domain, function (err, res) {
        if (err) {
            log.error(err['message']);
            if (cb) cb();
            return;
        };

        if (res && res['variables']) {
            var variables = res['variables'],
                _arr = [];
            for (var key in variables) {
                //if (typeof variables[key] === 'string') {
                _arr.push(key + '=' + variables[key]);
                //};
            };
            scope.domainVariables = res['variables'];
            scope.domainVariablesRecordId = res['_id'];
            if (_arr.length > 0) {
                scope.__setVar({
                    "setVar": _arr
                }, cb);
                return;
            };
        };
        if (cb) cb();
    });
};

CallRouter.prototype.setDomainVariable = function (key, value) {
    this.updateDomainVariable = true;
    this.domainVariables[key] = value;
};

CallRouter.prototype.destroyLocalRegExpValues = function () {
    var scope = this;
    Object.keys(this.regCollection).forEach(function (key) {
        if (key != (COMMANDS.REGEXP + 0)) {
            delete scope.regCollection[key];
        }
    });
};

var moment = require('moment-timezone');
CallRouter.prototype.DateOffset = function() {
    if (!this.offset) {
        return new Date();
    };
    return new Date(moment().tz(this.offset).format('LLLL'));
};

CallRouter.prototype.setChnVar = function (name, value) {
    this.connection.channelData.addHeader(name, value);
};

CallRouter.prototype.getChnVar = function (name) {
    var _var = this.connection.channelData.getHeader('variable_' + name)
        || this.connection.channelData.getHeader(name)
            //|| this.getLocalVariable(name)
        || this.connection.channelData.getHeader(VARIABLES_MAP[name])
        || '';
    return _var ;
};

CallRouter.prototype.getGlbVar = function (name) {
    try {
        var _var = this.globalVar[name];
        return _var
            ? _var
            : ''
    } catch (e) {
        return '';
    }
};

CallRouter.prototype.year = function (param) {
    return this._DateParser(param || '', (this.DateOffset().getFullYear()), 9999);
};

CallRouter.prototype.yday = function (param) {
    var now = new this.DateOffset(),
        start = new Date(now.getFullYear(), 0, 0),
        diff = now - start,
        oneDay = 1000 * 60 * 60 * 24;
    return this._DateParser(param, (Math.floor(diff / oneDay)), 366);
};

CallRouter.prototype.mon = function (param) {
    return this._DateParser(param, (this.DateOffset().getMonth() + 1), 12);
};

CallRouter.prototype.mday = function (param) {
    return this._DateParser(param, this.DateOffset().getDate(), 31);
};

CallRouter.prototype.week = function (param) {
    return this._DateParser(param, this.DateOffset()._getWeek(), 53);
};

CallRouter.prototype.mweek = function (param) {
    return this._DateParser(param, (this.DateOffset()._getWeekOfMonth() + 1), 6);
};

CallRouter.prototype.wday = function (param) {
    return this._DateParser(param, (this.DateOffset().getDay() + 1), 7);
};

CallRouter.prototype.hour = function (param) {
    return this._DateParser(param, this.DateOffset().getHours(), 23);
};

CallRouter.prototype.minute = function (param) {
    return this._DateParser(param, this.DateOffset().getMinutes(), 59);
};

CallRouter.prototype.minute_of_day = function (param) {
    var now = this.DateOffset();
    return this._DateParser(param, (now.getHours() * 60 + now.getMinutes()), 1440);
};


function _toInt (str) {
    return str.split(':').reduce( (r, c, i) => {
        if (i == 0) {
            return +c * 10000
        } else if ( i == 1) {
            return r + (+c * 100)
        }
        return r + ( +c )
    } , 0);
};


CallRouter.prototype.time_of_day = function (param) {
    param = param || "";
    let times = param.split(','),
        offsetDate = this.DateOffset(),
        current = (offsetDate.getHours() * 10000) + (offsetDate.getMinutes() * 100) + offsetDate.getSeconds(),
        _t;

    for (let i = 0, len = times.length; i < len; i++) {
        _t = times[i].split('-').map( (a) => _toInt(a));
        if ((current >= _t[0] && current <= _t[1])) return true;
    };
    return false;
};

CallRouter.prototype._DateParser = function (param, datetime, maxVal) {
    param = param || '';
    var datetimes = param.replace(/\s/g, '').split(','),
        result = false;
    if (datetimes[0] == "") {
        throw Error("bad parameters");
    };

    for (var i = 0; i < datetimes.length; i++) {
        log.trace('Offset: %s, expressionValue: %s, offsetValue: %s', this.offset, datetimes[i], datetime);
        result = (datetimes[i].indexOf('-') == -1)
            ? datetime == parseInt(datetimes[i])
            : equalsRange(datetime, datetimes[i], maxVal);

        if (result == true) {
            return result
        };
    };
    return result;
};

Date.prototype._getWeek = function() {
    var onejan = new Date(this.getFullYear(),0,1);
    return Math.ceil((((this - onejan) / 86400000) + onejan.getDay()+1)/7);
};

Date.prototype._getWeekOfMonth = function(exact) {
    var month = this.getMonth()
        , year = this.getFullYear()
        , firstWeekday = new Date(year, month, 1).getDay()
        , lastDateOfMonth = new Date(year, month + 1, 0).getDate()
        , offsetDate = this.getDate() + firstWeekday - 1
        , index = 1 // start index at 0 or 1, your choice
        , weeksInMonth = index + Math.ceil((lastDateOfMonth + firstWeekday - 7) / 7)
        , week = index + Math.floor(offsetDate / 7)
        ;
    if (exact || week < 2 + index) return week;
    return week === weeksInMonth ? index + 5 : week;
};

CallRouter.prototype.match = function (reg, val) {
    var _reg;
    _reg = reg.replace(/\u0001/g, '\\');
    _reg = _reg.match(new RegExp('^/(.*?)/([gimy]*)$'));
    if (!_reg) {
        _reg = [null, reg];
    };
    var _result = new RegExp(_reg[1], _reg[2]).exec(val);
    var _regOb = {};
    // TODO оптимизировать
    for (var key in _result) {
        _regOb['$' + key] = _result[key];
    };
    this.regCollection[COMMANDS.REGEXP + (Object.keys(this.regCollection).length)] = _regOb;
    return _result ? true : false;
};

CallRouter.prototype.setDestinationNumber = function (strReg, number) {
    if ( !(typeof strReg === 'string') ) return;
    var _tmp = strReg.match(new RegExp('^/(.*?)/([gimy]*)$'));
    if (!_tmp) {
        _tmp = [null, strReg];
    };
    var _reg = new RegExp(_tmp[1], _tmp[2]).exec(number);
    var _regOb = {};
    // TODO оптимизировать
    for (var key in _reg) {
        _regOb['$' + key] = _reg[key];
    };
    this.regCollection[COMMANDS.REGEXP + '0'] = _regOb;
};

CallRouter.prototype.__switch = function (condition, cb) {
    condition = condition[OPERATION.SWITCH];

    var _var = condition['variable'] || '',
        _case = condition['case'] || {},
        _value = this._parseVariable(_var);

    if (_case.hasOwnProperty(_value)) {
        log.trace('Switch %s = %s', _var, _value);
        this.execute(_case[_value], cb);
    } else if (_case.hasOwnProperty('default')) {
        log.trace('Switch %s = default', _var);
        this.execute(_case['default'], cb);
    } else {
        cb();
    };
};

CallRouter.prototype.__if = function (condition, cb) {
    condition = condition[OPERATION.IF];
    var sandbox = {
        _resultCondition: false//,
        // sys: this
    };
    if (condition['sysExpression']) {
        var expression = condition['sysExpression'] || '';

        log.trace('Parse expression: %s', expression);
        // TODO переделать на новый процесс.
        /* try {
         var script = vm.createScript('try { _resultCondition = (' + expression + ') } catch (e) {}');
         script.runInNewContext(sandbox);
         } catch (e) {
         log.error(e.message);
         };*/

        try {
            var _fn = new Function('sys, module, process', 'try { return (' + expression + ') } catch (e) {}');
            sandbox._resultCondition = _fn(this);
        } catch (e) {
            log.error(e);
        }
        log.trace('Condition %s : %s', expression, sandbox._resultCondition
            ? true
            : false);
        if (sandbox._resultCondition) {
            if (condition[OPERATION.THEN]) {
                this.execute(condition[OPERATION.THEN], cb);
            } else {
                cb();
            };
        } else {
            if (condition[OPERATION.ELSE]) {
                this.execute(condition[OPERATION.ELSE], cb);
            } else {
                cb();
            };
        };
    } else {
        log.error('Bad request IF');
        if (cb)
            cb();
    };
};

CallRouter.prototype._parseVariable = function (name) {
    var scope = this;
    name = name || '';
    try {
        return name
            .replace(/\$\$\{([\s\S]*?)\}/gi, function (a, b) {
                return scope.getGlbVar(b);
            })
            // ChannelVar
            .replace(/\$\{([\s\S]*?)\}/gi, function (a, b) {
                return scope.getChnVar(b)
            });
    } catch (e) {
        log.error('_parseVariable: %s', e.message);
    };
};

CallRouter.prototype.execApp = function (_obj, cb) {
    if (_obj[OPERATION.APPLICATION]) {
        if (typeof _obj[OPERATION.DATA] === 'string') {
            var scope = this;
            _obj[OPERATION.DATA] = _obj[OPERATION.DATA].replace(/\&reg(\d+)\.(\$\d+)/g, function(a, reg, key) {
                var _res = (scope.regCollection[COMMANDS.REGEXP + reg])
                    ? scope.regCollection[COMMANDS.REGEXP + reg][key]
                    : '';
                return _res || '';
            });
        };
        if (_obj[OPERATION.ASYNC]) {
            log.trace('Execute sync app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA] || '');
            this.connection.setEventLock(false);
        } else {
            this.connection.setEventLock(true);
            log.trace('Execute app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA] || '');
        };
        this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA] || '', cb);
    };
};

function getFnName(cond) {
    var propKeys = Object.keys(cond);
    if (propKeys.length === 1) {
        return propKeys[0];
    } else if (propKeys.length === 0) {
        return null;
    } else {
        for (var i = 0, len = propKeys.length; i < len; i++) {
            if (propKeys[i] != 'break' && propKeys[i] != 'async' && propKeys[i] != 'tag') {
                return propKeys[i];
            };
        };
    };
};

CallRouter.prototype.doExec = function (condition, cb) {
    if (condition instanceof Object) {

        if (this.versionSchema === 2) {

            var fnName = getFnName(condition);
            if (fnName && typeof this['__' + fnName] === 'function') {
                log.debug('execute application __' + fnName);
                this['__' + fnName](condition, cb);
            } else {
                log.error('Bad application %s.', fnName);
                if (cb) cb();
            };
        } else {
            if (condition.hasOwnProperty(OPERATION.IF)) {
                this.__if(condition, cb);
            } else if (condition.hasOwnProperty(OPERATION.APPLICATION)) {
                this.execApp(condition);
                if (cb) {
                    cb();
                };
            } else {
                if (cb) {
                    cb();
                };
                log.error('error parse json');
            };
        };
    };
};

CallRouter.prototype.__calendar = calendar;


CallRouter.prototype.execute = function (callflows, cb) {
    var scope = this;
    var i = 0;

    var postExec = function (err, res) {
        i++;
        scope.end = (callflows[i - 1] && callflows[i - 1] instanceof  Object &&
            callflows[i - 1]['break'] === true) || scope.end;
        if (i == callflows.length || scope.end) {
            if (cb)
                cb();
            return;
        };
        scope.doExec(callflows[i], postExec);
    };

    if (callflows instanceof Array && callflows.length > 0) {
        //callflows.forEach(function (callflow) {
        //    scope.doExec(callflow);
        //});

        try {
            this.doExec(callflows[i], postExec);
        } catch (e) {
            log.error(e);
        }
    };
};

CallRouter.prototype.run = function (callflows) {
    var scope = this;
    this.setupDomainVariables(function () {
        // TODO add scope.index = 0 (callback !!!!;)
        scope.start(callflows);
    });
};

CallRouter.prototype.stop = function () {
    this.end = true;
};

CallRouter.prototype.start = function (callflows) {
    this.callflows = callflows;
    var scope = this;

    var postExec = function (err, res) {
        scope.index++;
        scope.end = (callflows[scope.index - 1] && callflows[scope.index - 1] instanceof  Object &&
            callflows[scope.index - 1]['break'] === true) || scope.end;

        if (scope.index == callflows.length || scope.end) {
            //scope.updateLocalVariables();
            scope.saveDomainVariables();
            scope.connection.disconnect();
            return;
        }
        scope.doExec(callflows[scope.index], postExec);
    };

    if (callflows instanceof Array && callflows.length > 0) {
        //callflows.forEach(function (callflow) {
        //    scope.destroyLocalRegExpValues();
        //    scope.execute([callflow]);
        //});
        try {
            this.execute([callflows[this.index]], postExec);
        } catch (e) {
            log.error(e);
        }
    }
};

CallRouter.prototype.__answer = function (app, cb) {
    var _app;
    if (app[OPERATION.ANSWER] == "" || /\b200\b|\bOK\b/i.test(app[OPERATION.ANSWER])) {
        _app = FS_COMMAND.ANSWER;
    } else if (/\b183\b|\bSession Progress\b/i.test(app[OPERATION.ANSWER])) {
        _app = FS_COMMAND.PRE_ANSWER;
    } else if (/\b180\b|\bRinging\b/i.test(app[OPERATION.ANSWER])) {
        _app = FS_COMMAND.RING_READY;
    };
    if (_app) {
        this.execApp({
            "app": _app,
            "async": app[OPERATION.ASYNC] ? true : false
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.ANSWER]);
    };

    if (cb)
        cb();
};

CallRouter.prototype.__park = function (app, cb) {
    var _data = '',
        _auto = '',
        _app = app[OPERATION.PARK]
        ;
    if (!_app['name'] || !_app['lot']) {
        log.warn('Bad parameters park: name and lot required.');
        if (cb)
            cb();
        return;
    };

    if (_app['auto']) {
        _auto = 'auto ' + _app['auto'] + ' ';
        _app['lot'] = _app['lot'].replace('-', ' ');
    };

    _data = _data.concat(_app['name'], '@${domain_name} ', _auto, _app['lot']);

    this.execApp({
        "app": FS_COMMAND.PARK,
        "data": _data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();

};

CallRouter.prototype._addVariableArrayToChannelDump = function (variables) {
    if (variables instanceof  Array) {
        let scope = this;
        variables.forEach(function(variableStr) {
            if (typeof variableStr != 'string') return;
            let id = variableStr.indexOf('='),
                varName = variableStr.substring(0, id),
                varVal = variableStr.substring(id + 1, variableStr.length);
            scope.setChnVar('variable_' + varName, varVal);

            //fix public
            if (varName == 'default_language' && varVal == 'ru') {
                scope.connection.execute('set', 'sound_prefix=\/$${sounds_dir}\/ru\/RU\/elena');
            }
        });
    }
};

CallRouter.prototype.__setArray = function (app, cb) {
    var prop = app[OPERATION.TAGS],
        tagName = app['tagName'] || "webitel_tags",
        async = app[OPERATION.ASYNC] ? true : false;

    if (prop instanceof Array) {
        var scope = this;
        prop.forEach(function (item, index) {
            var _stIndex = isFinite(scope._dumpArrayIndex[tagName]) ? scope._dumpArrayIndex[tagName] : 0;
            scope.setChnVar('variable_' + tagName + '[' + (_stIndex++) + ']', (scope._parseVariable(item) || item));
            scope._dumpArrayIndex[tagName] = _stIndex;
            scope.execApp({
                "app": FS_COMMAND.PUSH,
                "data": tagName + ',' + item,
                "async": async
            });
        });
    }
    else if (prop instanceof Object) {
        var scope = this;
        for (let tag in prop) {
            if (prop.hasOwnProperty(tag) && prop[tag] instanceof Array) {
                var _stIndex = isFinite(scope._dumpArrayIndex[tag]) ? scope._dumpArrayIndex[tag] : 0;
                prop[tag].forEach((item, index) =>{
                    scope.setChnVar('variable_' + tag + '[' + (_stIndex++) + ']', (scope._parseVariable(item) || item));
                    scope.execApp({
                        "app": FS_COMMAND.PUSH,
                        "data": tag + ',' + item,
                        "async": async
                    });
                });
                scope._dumpArrayIndex[tag] = _stIndex;
            }
        }
    } else if (typeof prop === 'string') {
        var _stIndex = isFinite(scope._dumpArrayIndex[tagName]) ? scope._dumpArrayIndex[tagName] : 0;
        scope.setChnVar('variable_' + tagName + '[' + (_stIndex++) +']', (scope._parseVariable(prop) || prop));
        scope._dumpArrayIndex[tagName] = _stIndex;
        this.execApp({
            "app": FS_COMMAND.PUSH,
            "data": tagName + ',' + prop,
            "async": async
        });
    }
    else {
        log.warn('Bad parameters __setTags');
    };

    if (cb)
        cb();

};

CallRouter.prototype.__setVar = function (app, cb) {
    var _app, _data, _chnArrayVar = [];

    if (app[OPERATION.SET] instanceof Array) {
        _app = FS_COMMAND.MULTISET;
        _data = '^^~' + app[OPERATION.SET].join('~');
        _chnArrayVar = app[OPERATION.SET];
    }
    else if (app[OPERATION.SET] instanceof Object) {
        var prop = app[OPERATION.SET];
        _data = [];
        for (var key in prop['data']) {
            if (prop['data'].hasOwnProperty(key)) {
                _chnArrayVar.push(key + '=' + prop['data'][key]);
                _data.push((prop['type'] == 'nolocal' ? 'nolocal:' + key : key) + '=' + prop['data'][key]);

                if (prop['type'] == 'domain') {
                    this.setDomainVariable(key, prop['data'][key]);
                };
            };
        };
        switch (prop['type']) {
            case 'all':
                _app = FS_COMMAND.EXPORT;
                break;
            case 'nolocal':
                _chnArrayVar = null;
                _app = FS_COMMAND.EXPORT;
                break;
            default :
                if (_data.length > 1) {
                    _app = FS_COMMAND.MULTISET;
                    _data = '^^~' + _data.join('~')
                } else {
                    _app = FS_COMMAND.SET;
                    _data = _data[0]
                }
        };
    }
    else {
        if (app[OPERATION.SET].indexOf('all:') == 0) {
            _app = FS_COMMAND.EXPORT;
            _data = app[OPERATION.SET].substring(4);
            _chnArrayVar = [_data];
        } else if (app[OPERATION.SET].indexOf('nolocal:') == 0) {
            _app = FS_COMMAND.EXPORT;
            _data = app[OPERATION.SET];
        } else if (app[OPERATION.SET].indexOf('domain:') == 0) {
            var tmpStr = app[OPERATION.SET].substring(7),
                tmp = tmpStr.split('=');
            this.setDomainVariable(tmp[0], tmp[1] || '');
            _app = FS_COMMAND.SET;
            _data = tmpStr;
            _chnArrayVar = [_data];
        } else {
            _app = FS_COMMAND.SET;
            _data = app[OPERATION.SET];
            _chnArrayVar = [_data];
        };
    };

    if (_app) {
        if (_data instanceof Array){
            for (var i = 0, len = _data.length; i < len; i++) {
                this.execApp({
                    "app": _app,
                    "data": _data[i],
                    "async": app[OPERATION.ASYNC] ? true : false
                });
            };
        } else {
            this.execApp({
                "app": _app,
                "data": _data,
                "async": app[OPERATION.ASYNC] ? true : false
            });
        }
    } else {
        log.warn('Bad parameter ', app[OPERATION.SET]);
    };

    this._addVariableArrayToChannelDump(_chnArrayVar);

    if (cb)
        cb();
};

function _getGotoDataString(param) {
    param = param || '';
    if (param.indexOf('default:') == 0) {
        return param.substring(8) + ' XML default';
    } else if (param.indexOf('public:') == 0) {
        return param.substring(7) + ' XML public';
    } else {
        return param;
    };
};


CallRouter.prototype.__break = function (app, cb) {
    cb();
};

function findPositionByTagName (callFlows, tagName) {
    try {
        for (let i = 0, len = callFlows.length; i < len; i++) {
            if (callFlows[i]['tag'] === tagName) return i;
        };
    } catch(e) {
        log.error(e);
    }
}

CallRouter.prototype.__goto = function (app, cb) {
    var _app = FS_COMMAND.TRANSFER;
    if (app[OPERATION.GOTO] && app[OPERATION.GOTO].indexOf('local:') === 0) {
        let _gotoIndexName = app[OPERATION.GOTO].substring(6);
        var _i = isFinite(+_gotoIndexName) ? parseInt(_gotoIndexName) : findPositionByTagName(this.callflows, _gotoIndexName);
        if (!isNaN(_i) && this.index !== _i) {
            log.trace('GOTO ' + _i);
            this.index = _i;
            if (++this.cycleCount === MAX_CYCLE_COUNT) {
                throw 'Cycle max count';
            };
            this.start(this.callflows);
        } else {
            log.error('Command "goto" cycle!');
            if (cb)
                cb();
        };
        return;
    };
    var _data = _getGotoDataString(app[OPERATION.GOTO]);

    this.execApp({
        "app": _app,
        "data": _data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};

CallRouter.prototype.__recordFile = function (app, cb) {
    let prop = app[OPERATION.RECORD_FILE] || {},
        name = encodeURI(prop['name'] || "recordFile"),
        playbackTerminators = prop['terminators'] || "#",
        type = prop['type'] || "mp3",
        maxSec = parseInt(prop['maxSec']) || 60,
        silenceThresh = parseInt(prop['silenceThresh']) || 200,
        silenceHits = parseInt(prop['silenceHits']) || 5,
        email = prop['email'] instanceof Array ? prop['email'].join(',') : 'none'
    ;

    var prevRecordFile = this.getChnVar(WEBITEL_RECORD_FILE_NAME);
    if (prevRecordFile) {
        this.execApp({
            "app": FS_COMMAND.STOP_RECORD_SESSION,
            "data": "/recordings/" + prevRecordFile
        });
    };

    let multiSet = '^^~playback_terminators=' + playbackTerminators
        + '~record_post_process_exec_api=luarun:RecordUpload.lua ${uuid} ${domain_name} ' + type + ' ' + email + ' ' + name;

    this.execApp({
        "app": FS_COMMAND.MULTISET,
        "data": multiSet
    });

    this.execApp({
        "app": FS_COMMAND.RECORD,
        "data": "/recordings/${uuid}_" + name +  "." + type + ' ' + maxSec + ' ' + silenceThresh + ' ' + silenceHits
    });

    return cb && cb();

};

CallRouter.prototype.__recordSession = function (app, cb) {
    var prop = app[OPERATION.RECORD_SESSION];
    var action,
        type,
        name = encodeURI(prop['name'] || 'recordSession'),
        email = 'none';

    if (typeof prop == 'string'){
        action = prop === 'stop' ? 'stop' : 'start';
        type = 'mp3';
    } else if (typeof prop == 'object') {
        action = prop['action'] === 'stop' ? 'stop' : 'start';
        type = prop['type'] === 'mp4' ? 'mp4' : 'mp3';
        email = prop['email'] instanceof Array ? prop['email'].join(',') : 'none';
    } else {
        log.error('Bad request __recordSession');
        if (cb)
            cb();
        return;
    };

    let varFileName = "${uuid}_" + name + "." + type;

    if (action == 'start') {
        this.__setVar({
            "setVar": WEBITEL_RECORD_FILE_NAME + '=' + varFileName
        });

        var multiSet = '^^~RECORD_MIN_SEC=' + (prop['minSec'] || '2' )
            + '~RECORD_STEREO=' + (String(prop['stereo']) == 'false' ? 'false' : 'true')
            + '~RECORD_BRIDGE_REQ=' + (String(prop['bridged']) == 'false' ? 'false' : 'true')
            + '~recording_follow_transfer=' + (String(prop['followTransfer']) == 'false' ? 'false' : 'true')
            + '~record_post_process_exec_api=luarun:RecordUpload.lua ${uuid} ${domain_name} ' + type + ' ' + email + ' ' + name;

        this.execApp({
            "app": FS_COMMAND.MULTISET,
            "data": multiSet
        });

        this.execApp({
            "app": FS_COMMAND.RECORD_SESSION,
            "data": "/recordings/" + varFileName
        });
    }
    else if (action == 'stop') {
        this.execApp({
            "app": FS_COMMAND.STOP_RECORD_SESSION,
            "data": "/recordings/" + varFileName
        });
    } else {
        log.warn('Bad parameters ', prop);
    };

    if (cb)
        cb();
};

CallRouter.prototype.__hangup = function (app, cb) {
    this.execApp({
        "app": FS_COMMAND.HANGUP,
        "data": app[OPERATION.HANGUP] || '',
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};

CallRouter.prototype.__script = function (app, cb) {
    var _data = 'lua/',
        _app = FS_COMMAND.LUA,
        prop = app[OPERATION.SCRIPT];

    if (prop instanceof Object && prop.hasOwnProperty('name')) {
        if (prop['type'] == 'js') {
            _app = FS_COMMAND.JS;
            _data = 'js/';
        };

        _data = _data.concat(prop['name']);

        if (prop['parameters'] instanceof Array) {
            for (var i = 0, len = prop['parameters'].length; i < len; i++) {
                _data = _data.concat(' "', prop['parameters'][i], '"');
            };
        };

        this.execApp({
            "app": _app,
            "data": _data,
            "async": app[OPERATION.ASYNC] ? true : false
        });

    } else {
        log.warn('Bad script name.');
    };

    if (cb)
        cb();
};

CallRouter.prototype.__log = function (app, cb) {
    if (typeof app[OPERATION.LOG] == 'string') {
        this.execApp({
            "app": FS_COMMAND.LOG,
            "data": 'CONSOLE ' + app[OPERATION.LOG],
            "async": app[OPERATION.ASYNC] ? true : false
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.SCRIPT]);
        return false;
    };

    if (cb)
        cb();
};

CallRouter.prototype.__echo = function (app, cb) {
    var _app, _data = '', delay = parseInt(app[OPERATION.ECHO]);

    if (!delay)
        delay = +this._parseVariable(app[OPERATION.ECHO]);

    if (delay > 0) {
        _app = FS_COMMAND.DELAY_ECHO;
        _data = app[OPERATION.ECHO];
    } else {
        _app = FS_COMMAND.ECHO;
    };
    this.execApp({
        "app": _app,
        "data": _data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};

// TODO set var from cb...
CallRouter.prototype.__httpRequest = function (app, cb) {
    httpReq(app[OPERATION.HTTP], this, cb);
};

CallRouter.prototype.__sendSms = function (app, cb) {
    sms(app[OPERATION.SEND_SMS], this, cb);
};

CallRouter.prototype.__sleep = function (app, cb) {
    var delay = parseInt(app[OPERATION.SLEEP]);
    if (!delay)
        delay = this._parseVariable(app[OPERATION.SLEEP]);

    this.execApp({
        "app": FS_COMMAND.SLEEP,
        "data": delay,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

CallRouter.prototype.__conference = function (app, cb) {
    var _data = '', prop = app[OPERATION.CONFERENCE];
    if (prop['name'] /*&& /^[a-zA-Z0-9+_-]+$/.test(prop['name'] )*/) {
        _data = _data.concat(prop['name'], '_', this.domain, '@',
            prop.hasOwnProperty('profile') ? prop['profile'] : 'default',
            prop.hasOwnProperty('pin') ? '+' + prop['pin'] : ''
        );

        if (prop.hasOwnProperty('flags') && prop['flags'] instanceof Array) {
            _data += '+flags{' + prop['flags'].join('|') + '}';
        };

        this.execApp({
            "app": FS_COMMAND.CONFERENCE,
            "data": _data,
            "async": app[OPERATION.ASYNC] ? true : false
        });
    } else {
        log.warn("Conference name ASCII letters, _, +, - or numbers ");
    };

    if (cb)
        cb();
};

CallRouter.prototype.__schedule = function (app, cb) {
    var _data = '+',
        _app,
        prop = app[OPERATION.SCHEDULE];

    _data += isNaN(prop['seconds'])
        ? '0 '
        : prop['seconds'] + ' ';

    if (prop['action'] == OPERATION.HANGUP) {
        _app = FS_COMMAND.SCHEDULE_HANGUP;
        _data = _data.concat(prop['data'] ? prop['data'] : '');
    } else if (prop['action'] == OPERATION.GOTO) {
        _app = FS_COMMAND.SCHEDULE_TRANSFER;

        _data = _data.concat( _getGotoDataString( prop['data']
            ? prop['data']
            : ''));
    } else {
        log.warn("Bad parameters SCHEDULE");
    };

    if (_app)
        this.execApp({
            "app": _app,
            "data": _data,
            "async": prop[OPERATION.ASYNC] ? true : false
        });

    if (cb)
        cb();
};

// TODO delete type, fileName, refresh
CallRouter.prototype._getPlaybackFileString = function (type, fileName, refresh, noPref, allProp = {}) {
    var filePath = '';
    fileName = this._parseVariable(fileName);
    // TODO delete
    log.debug('Parse playback file -> %s', fileName);

    switch (type) {
        case MEDIA_TYPE.WAV:
            var cdrUrl = this.getGlbVar('cdr_url');
            if (cdrUrl) {
                filePath = (refresh === true ? '{refresh=true}' : '') + "http_cache://" +
                    encodeURI(cdrUrl + '/sys/media/' + MEDIA_TYPE.WAV + '/' + fileName + '?stream=false&domain=' + this.domain + '&.wav');
            };
            break;
        case MEDIA_TYPE.LOCAL:
            filePath = fileName;
            break;
        case MEDIA_TYPE.SILENCE:
            filePath = noPref ? type : 'silence_stream://' + fileName;
            break;
        case MEDIA_TYPE.SHOUT:
            filePath = (fileName || '').replace(/https?/, 'shout');
            break;
        case MEDIA_TYPE.TONE:
            filePath = noPref ? fileName : 'tone_stream://' + fileName;
            break;
        case MEDIA_TYPE.SAY:

            let [lang = "en", method = "number pronounced"] = [allProp.lang, allProp.method];
            lang = this._parseVariable(lang);
            filePath = "${" + `say_string ${lang} ${lang} ${method} ${fileName}` + "}";
            break;
        default :
            var cdrUrl = this.getGlbVar('cdr_url');
            if (cdrUrl) {
                filePath = encodeURI(cdrUrl.replace(/https?/, 'shout') + '/sys/media/' + MEDIA_TYPE.MP3 + '/' + fileName
                    + '?domain=' + this.domain);
            };
    };

    return filePath;
};

CallRouter.prototype.__playback = function (app, cb) {
    var filePath = '',
        prop = app[OPERATION.PLAYBACK],
        broadcast = prop['broadcast'],
        _terminator = prop['terminator'],
        scope = this;

    if (typeof prop['name'] === 'string') {
        filePath = this._getPlaybackFileString(prop['type'], prop['name'], prop['refresh'], false, prop);
    } else if (prop['files'] instanceof Array) {
        var files = prop['files'];
        for (var i = 0, len = files.length; i < len; i++) {
            filePath += '!' + this._getPlaybackFileString(files[i]['type'], files[i]['name'], files[i]['refresh'], false, files[i]);
        };
        filePath = 'file_string://' + filePath.substring(1);
    } else {
        log.warn('Bad _playback parameters');
        if (cb)
            cb();
        return;
    };

    if (app[OPERATION.PLAYBACK].hasOwnProperty('getDigits')) {
        var _playAndGetDigits = app[OPERATION.PLAYBACK]['getDigits'],
            _setVar = _playAndGetDigits['setVar'] || 'MyVar',
            _min = _playAndGetDigits['min'] || 1,
            _max = _playAndGetDigits['max'] || 1,
            _tries = _playAndGetDigits['tries'] || 1,
            _timeout = _playAndGetDigits['timeout'] || 3000;

        this.execApp({
            "app": FS_COMMAND.PLAY_AND_GET,
            "data": [_min, _max, _tries, _timeout, _terminator || '#', filePath, 'silence_stream://250', _setVar, '\\d+'].join(' '),
            "async": app[OPERATION.PLAYBACK][OPERATION.ASYNC] ? true : false
        }, function (res) {
            try {
                var _r = res.getHeader('variable_' + _setVar) || '';
                scope._addVariableArrayToChannelDump([_setVar + '=' + _r]);
                log.trace('Set %s = %s', _setVar, _r);
            } catch (e) {
                log.error(e.message);
            }
            if (cb)
                cb();
        });

    } else {
        if (broadcast) {
            if (!~['aleg', 'bleg', 'both'].indexOf(broadcast))
                broadcast = 'both';

            this.connection.bgapi(`uuid_broadcast ${this.uuid} ${filePath} ${broadcast}`);
        } else {
            if (_terminator) {
                this.execApp({
                    "app": FS_COMMAND.SET,
                    "data": "playback_terminators=" + _terminator
                });
            }

            this.execApp({
                "app": FS_COMMAND.PLAYBACK,
                "data": filePath,
                "async": app[OPERATION.PLAYBACK][OPERATION.ASYNC] ? true : false
            });
        }

        if (cb)
            cb();
    }
};

CallRouter.prototype.__bridge = function (app, cb) {
    var prop = app[OPERATION.BRIDGE],
        _data = '',
        scope = this,
        separator = prop['strategy'] == 'failover' // TODO переделать
            ? '|'
            : ','; // ":_:" - only for user & device; "," - for other types

    if (prop.hasOwnProperty('global') && prop['global'] instanceof Array){
        _data += '<' + prop['global'].join(',') + '>';
    };
    _data += '{' + 'domain_name=' + this.domain;
    if (prop.hasOwnProperty('parameters') && prop['parameters'] instanceof Array) {
        _data = _data.concat(',', prop['parameters'].join(','));
    };
    _data += '}';

    if (prop.hasOwnProperty('endpoints') && prop['endpoints'] instanceof Array) {
        prop['endpoints'].forEach(function (endpoint) {
            switch (endpoint['type']) {
                case 'sipGateway':
                    if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                        _data = _data.concat('[', endpoint['parameters'].join(','), ']')
                    };
                    _data = _data.concat('sofia/gateway/', endpoint['name'], '/', endpoint['dialString']);
                    break;
                case 'sipUri':
                    if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                        _data = _data.concat('[', endpoint['parameters'].join(','), ']')
                    };
                    _data = _data.concat('sofia/', endpoint.hasOwnProperty('profile') ? endpoint['profile'] : 'external',
                        '/', endpoint['dialString'], '@', endpoint['host']);
                    break;
                case 'sipDevice':
                    if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                        _data = _data.concat('[', endpoint['parameters'].join(','), ']')
                    };
                    _data = _data.concat('sofia/', endpoint.hasOwnProperty('profile') ? endpoint['profile'] : 'external',
                        '/', endpoint['name'], '%', endpoint['domainName'], '^', endpoint['dialString']);
                    break;
                case 'device':

                    scope.execApp({
                        "app": FS_COMMAND.HASH,
                        "data": "insert/spymap/${domain_name}-" + endpoint['name'] +  "/${uuid}"
                    });

                    if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                        _data = _data.concat('[', endpoint['parameters'].join(','), ']')
                    };
                    _data = _data.concat('user/', endpoint['name'], '@${domain_name}');
                    break;
                case 'user':
                    //TODO move to fn
                    scope.execApp({
                        "app": FS_COMMAND.HASH,
                        "data": "insert/spymap/${domain_name}-" + endpoint['name'] +  "/${uuid}"
                    });

                    switch (endpoint['proto']) {
                        case "sip":
                            _data = _data.concat('[', 'webitel_call_uuid=${create_uuid()},sip_invite_domain=${domain_name},' +
                                'presence_id=', endpoint['name'], '@${domain_name}');
                            if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                                _data = _data.concat(',', endpoint['parameters'].join(','));
                            };
                            _data = _data.concat(']${sofia_contact(*/', endpoint['name'],'@${domain_name})}');
                            break;
                        case "webrtc":
                            _data = _data.concat('[', 'webitel_call_uuid=${create_uuid()},sip_invite_domain=${domain_name},' +
                                'presence_id=', endpoint['name'], '@${domain_name}');
                            if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                                _data = _data.concat(',', endpoint['parameters'].join(','));
                            };
                            _data = _data.concat(']${verto_contact(', endpoint['name'], '@${domain_name})}');
                            break;
                        default :
                            if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                                _data = _data.concat('[', endpoint['parameters'].join(','), ']')
                            };
                            _data = _data.concat('user/', endpoint['name'], '@', endpoint.hasOwnProperty('domainName')
                                ? endpoint['domainName']
                                : '${domain_name}');
                            break;
                    };
                    break;
            };
            _data = _data.concat(separator);
        });

        var pickup = '';
        if (prop['pickup'] && prop['strategy'] != 'failover') {
            if (prop['pickup'] instanceof Array) {
                prop['pickup'].forEach(function (item) {
                    pickup += ',pickup/' + item + '@${domain_name}'
                });
            } else {
                pickup += ',pickup/' + prop['pickup'] + '@${domain_name}'
            };
        }

        // TODO WTEL-263
        this.execApp({
            "app": FS_COMMAND.BRIDGE,
            "data": _data.slice(0, (-1 * separator.length)) + pickup,
            "async": prop[OPERATION.ASYNC] ? true : false
        }, function (res) {
            res.headers.forEach( (item) => {
                //console.log('set: ' + item.name + ' => ' + item.value);
                this.channelData.addHeader(item.name, item.value);
            });
            // TODO continue_on_fail=?
            //
            if (res.getHeader('variable_bridge_hangup_cause') === 'NORMAL_CLEARING' && res.getHeader('variable_hangup_after_bridge') === 'true') {
                scope.stop();
                if (cb)
                    cb();
            } else {
                if (cb)
                    cb();
            }

        });

    };
};

CallRouter.prototype.__inBandDTMF = function (app, cb) {
    var application = app[OPERATION.IN_BAND_DTMF] == 'stop' ? FS_COMMAND.STOP_DTMF : FS_COMMAND.START_DTMF;
    this.execApp({
        "app": application,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

CallRouter.prototype.__flushDTMF = function (app, cb) {
    this.execApp({
        "app": FS_COMMAND.FLUSH_DTMF,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

class RouterTimer {
    constructor (option, router) {
        if (!(option['actions'] instanceof Array)) {
            //TODO log error
            log.error('RouterTimer: bad parameters');
            return
        };

        this.tries = option['tries'] || Infinity;
        this.offset = (option['offset'] * 1000) || 0;
        this.interval = (option['interval'] || 60) * 1000;
        this._tries = 1;
        this._stop = false;

        var scope = this;

        this._timerId = setTimeout( function tick() {

                router.execute(option['actions'], () => {

                    if (scope._stop || ++scope._tries > scope.tries)
                        return;

                    console.log(`_tries: ${scope._tries}; interval: ${scope.interval}; offset: ${scope.offset}; tries: ${scope.tries};`);

                    if ((scope.interval += scope.offset) < 1000) {
                        log.error('Bad time, interval less than 1');
                        return scope.stop();
                    };

                    scope._timerId = setTimeout(tick, scope.interval);
                });

        },  this.interval);
    }

    stop () {
        this._stop = true;
        return clearTimeout(this._timerId);
    }
}

CallRouter.prototype.__queue = function (app, cb) {
    var _data = '',
        prop = app[OPERATION.QUEUE],
        queueName = prop['name'],
        continueOnAnswered = prop.continueOnAnswered
        ;

    if (typeof queueName != 'string') {
        log.error('Bad parameters queue.');
        if (cb)
            cb();
        return;
    };

    if (queueName.indexOf('$') == 0) {
        queueName = this._parseVariable(queueName)
    };


    if (queueName && /^[a-zA-Z0-9+_-]+$/.test(queueName)) {
        _data = queueName + '@${domain_name}';
    } else {
        log.error('Bad parameters queue name.');
        if (cb)
            cb();
        return;
    };

    var scope = this,
        timer = prop['timer'],
        _removeListeners = false
        ;

    var _closeTimer = function () {
        ccTimers.forEach( (item) => item.stop() );
        ccTimers.length = 0;
    };

    var ccTimers = [];

    if (timer instanceof Object) {
        if (timer instanceof Array) {

            timer.forEach( (t) => {
                ccTimers.push(new RouterTimer(t, scope));
            });

        } else {
            ccTimers.push(new RouterTimer(timer, scope));
        }

        this.connection.on('esl::end', _closeTimer);

        this.connection.on('error', _closeTimer);

        _removeListeners = true;
    };

    this._curentQueue = queueName + '@' + this.domain;

    this.execApp({
        "app": FS_COMMAND.CALLCENTER,
        "data": _data,
        "async": (app[OPERATION.ASYNC] ? true : false) || !!timer
    }, function(res) {
        if (!continueOnAnswered && res.getHeader('variable_cc_cause') == 'answered')
            app.break = true;

        scope._curentQueue = null;
        log.debug('Callback queue: %s', queueName);
        if (_removeListeners) {
            _closeTimer();
            scope.connection.removeListener('error', _closeTimer);
            scope.connection.removeListener('esl::end', _closeTimer);
        };
        if (cb)
            cb();
    });

    if (prop.hasOwnProperty('startPosition') && prop['startPosition']) {
        let _varName,
            _verbose = false,
            _fireEvent = false;
        if (prop['startPosition'] instanceof Object) {
            _varName = prop['startPosition']['var'] || 'cc_start_position';
            _verbose = !!prop['startPosition']['verbose'];
            _fireEvent = prop['startPosition']['event'];
        } else if (typeof prop['startPosition'] == 'string') {
            _varName = prop['startPosition'] || 'cc_start_position'
        } else {
            _varName = 'cc_start_position'
        }
        this.__ccPosition({
            "ccPosition": {
                "var": _varName
            }
        }, function () {
            if (_fireEvent) {
                let headers = {};
                headers[_varName] = '${' + _varName + '}';
                scope.__event({
                    "event": {
                        "action": typeof _fireEvent == 'string' ? _fireEvent : "cc_start_position",
                        "verbose": _verbose,
                        "headers": headers
                    }
                })
            }
        })
    }
};

CallRouter.prototype.__ccPosition = function (app, cb) {
    let prop = app[OPERATION.CC_POSITION],
        varName = prop && prop['var'],
        scope = this
    ;

    if (!varName) {
        log.error('Bad parameters ccPosition');
        if (cb)
            cb();
        return;
    };

    if (!this._curentQueue) {
        log.warn('Queue empty');
        if (cb)
            cb();
        return;
    }

    this.connection.api('callcenter_config queue list members ' + this._curentQueue, function (res) {
        try {
            let body = res.body || '';
            let position = 1;
            let lines = body.match(/[^\r\n]+/g);

            for (var line of lines) {
                if (line.indexOf('Trying') != -1 || line.indexOf('Waiting') != -1) {
                    if (line.indexOf(scope.uuid) != -1) {
                        break;
                    }
                    ;
                    position++;
                };
            };

            scope.__setVar({
                "setVar": varName + '=' + position
            }, function () {
                scope.__setVar({
                    "setVar": "cc_export_vars=" + varName
                });
                if (cb)
                    cb();
            });


        } catch (e) {
            log.error(e);
            if (cb)
                cb();
        }
    })
};

CallRouter.prototype.__setSounds = function (app, cb) {
    let prop = app[OPERATION.SET_SOUNDS] || {},
        voice = prop['voice'],
        lang = prop['lang'],
        soundPref
        ;

    if (typeof voice  != 'string' || typeof lang != 'string') {
        log.error('Bad parameters setLanguage');
        if (cb)
            cb();
        return;
    };

    let tmp = lang
        .toLowerCase()
        .split('_');

    soundPref = '\/$${sounds_dir}\/' + tmp.join('\/') + '\/' + voice;

    this.__setVar({
        "setVar": ['sound_prefix=' + soundPref, 'default_language=' + tmp[0]]
    }, function () {
        if (cb)
            return cb();
    });
};

CallRouter.prototype.__exportVars = function (app, cb) {
    var _item = {}, prop = app[OPERATION.EXPORT_VARS], scope = this;

    if (prop instanceof Array) {
        prop.forEach(function (item) {
            _item[item] = scope.getChnVar(item);
        });

        scope.__setVar({
            "setVar": 'webitel_data=' + JSON.stringify(_item)
        }, function () {
            scope.__setVar({
                "setVar": 'cc_export_vars=webitel_data'
            }, function () {
                if (cb)
                    return cb();
            });
        });
    } else {
        log.error('Bad __exportVars parameters.');
    };
};

CallRouter.prototype.__event = function (app, cb) {
    let prop = app[OPERATION.EVENT] || {},
        name = prop['action'],
        verbose = prop['verbose'],
        headers = prop['headers']
    ;

    if (!name) {
        log.error('Bad parameters application event');
        if (cb)
            cb();
        return;
    };

    if (verbose) {
        let data = ['Event-Subclass=webitel::acr', 'Event-Name=CUSTOM', `action=${name}`, 'domain=' + this.domain];
        for (let key in headers) {
            data.push(key + '=' + (this._parseVariable(headers[key]) || key));
        };

        this.execApp({
            "app": FS_COMMAND.EVENT,
            "data": data.join(','),
            "async": app[OPERATION.ASYNC] ? true : false
        });

    } else {
        let event = new Event('CUSTOM', 'webitel::acr');

        for (let key in headers) {
            event.addHeader(key, this._parseVariable(headers[key]) || key);
        };

        event.addHeader('action', name);
        event.addHeader('domain', this.domain);
        event.addHeader('type', 'text/plain');
        event.addHeader('Content-Type', 'text/plain');
        event.addBody("+OK");

        this.connection.sendEvent(event, (e, r) => {
            log.trace(`Send event "${name}": ${r["Reply-Text"]}`);
        });
    };

    if (cb)
        cb();
};

CallRouter.prototype.__ivr = function (app, cb) {
    if (typeof app[OPERATION.IVR] === 'string') {

        this.execApp({
            "app": FS_COMMAND.IVR,
            "data": app[OPERATION.IVR] + '@' + this.domain,
            "async": app[OPERATION.ASYNC] ? true : false
        });

    } else {
        log.error('Bar ivr menu parameters');
    }
    if (cb)
        cb();
};

CallRouter.prototype.__voicemail = function (app, cb) {
    var prop = app[OPERATION.VOICEMAIL] || {},
        domain = this['domain'],
        user = prop['user'] || ''
        ;

    if (prop['check'] === true) {

        if (typeof prop['announce'] === 'boolean') {
            this.execApp({
                "app": FS_COMMAND.SET,
                "data": "vm_announce_cid=" + prop['announce'].toString(),
                "async": false
            });
        };

        var auth = (typeof prop['auth'] === 'boolean')
                ? (!prop['auth']).toString()
                : "${sip_authorized}"
            ;

        this.execApp({
            "app": FS_COMMAND.SET,
            "data": 'voicemail_authorized=' + auth,
            "async": false
        });

        this.execApp({
            "app": FS_COMMAND.VOICEMAIL,
            "data": 'check default ' + domain + ' ' + user,
            "async": app[OPERATION.ASYNC] ? true : false
        }, function() {
            if (cb)
                cb();
        });

    } else {

        if (prop['user'] == '') {
            log.error('Bad voicemail parameters.');
            if (cb) cb(new Error('Bad voicemail parameters.'));
            return;
        };

        var _set = [];
        if (prop['skip_greeting'] === true)
            _set.push('skip_greeting=true');

        if (prop['skip_instructions'] === true)
            _set.push('skip_instructions=true');

        if (prop['cc'] instanceof Array) {
            var cc = '';
            prop['cc'].forEach(function (item, index) {
                if (index > 0)
                    cc += ',';

                cc += item + '@' + domain
            });
            _set.push('vm_cc=' + cc);
        };

        if (_set.length > 0) {
            this.__setVar({
                "setVar": _set
            });
        };

        this.execApp({
            "app": FS_COMMAND.VOICEMAIL,
            "data": 'default ' + domain + ' ' + prop['user'],
            "async": app[OPERATION.ASYNC] ? true : false
        });
        if (cb)
            cb();
    };
};

CallRouter.prototype.__bindAction = function (app, cb) {
    var prop = app[OPERATION.BIND_ACTION];

    if (!prop['action'] || !prop['name']) {
        log.error('Bad parameters bind_action');
        if (cb)
            cb();
        return;
    };

    var type = prop['type'] || 'exec';

    var data = prop['name'] + ',' + prop['digits'] + ',' + type + ':' + prop['action'];

    if (prop['parameters'] instanceof Array) {
        data += ',' + prop['parameters'].join(',');
    } else if (typeof prop['parameters'] === 'string') {
        data += ',' + prop['parameters'];
    };

    this.connection.execute('export', 'domain_name=' + this.domain);

    var scope = this;
    this.execApp({
        "app": FS_COMMAND.BIND_DIGIT_ACTION,
        "data": data,
        "async": prop[OPERATION.ASYNC] ? true : false
    }, function () {

    });

    scope.execApp({
        "app": 'digit_action_set_realm',
        "data": prop['name'],
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

CallRouter.prototype.__clearAction = function (app, cb) {
    if (typeof app[OPERATION.CLEAR_ACTION] !== 'string') {
        log.error('Bad parameters clear_action');
        if (cb)
            cb();
        return;
    };

    var data = app[OPERATION.CLEAR_ACTION] == ""
        ? "all"
        : app[OPERATION.CLEAR_ACTION];
    this.execApp({
        "app": FS_COMMAND.CLEAR_DIGIT_ACTION,
        "data": data,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

CallRouter.prototype.__bindExtension = function (app, cb) {
    var prop = app[OPERATION.BIND_EXTENSION];
    if (!prop['digits'] || !prop['digits']) {
        log.error('Bad parameters _bind_extension');
        if (cb)
            cb();
        return;
    };

    var listen = prop['listen'] || 'a';
    var respond = prop['respond'] || 's';

    var data = prop['digits'] + ' ' + listen + ' ' + respond + ' execute_extension::' + prop['extension'].toString();

    this.execApp({
        "app": FS_COMMAND.BIND_EXTENSION,
        "data": data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};

CallRouter.prototype.__attXfer = function (app, cb) {
    // findExtension
    var prop = app[OPERATION.ATT_XFER],
        scope = this;
    if (!prop['destination']) {
        log.error('Bar request _att_xfer');
        if (cb)
            cb();
        return;
    };

    var destination = this._parseVariable(prop['destination']) || prop['destination'];

    findExtension(destination, scope.domain, function (err, res) {
        if (err) {
            log.error(err['message']);
            if (cb)
                cb();
            return;
        };
        var data = '';
        if (res || !prop['gateway']) {
            data = 'user/' + destination + '@' + scope.domain;
        } else {
            data = 'sofia/gateway/' + prop['gateway'] + '/' + destination;
        };

        var caller_id_number = scope.getChnVar('Caller-ANI') || '';
        //console.log(scope.connection.channelData.serialize('plain'));
        scope.execApp({
            "app": FS_COMMAND.ATT_XFER,
            "data": '{webitel_direction=outbound,domain_name=' + scope.domain + ',effective_caller_id_name=' + caller_id_number +
            ',effective_caller_id_number=' + caller_id_number +'}' + data,
            "async": app[OPERATION.ASYNC] ? true : false
        });

        if (cb)
            cb();
    });
};

CallRouter.prototype.__unSet = function (app, cb) {
    if (typeof app[OPERATION.UN_SET] !== 'string') {
        log.error('bad request _unSet');
        if (cb) {
            cb();
        }
        return;
    }
    this.execApp({
        "app": FS_COMMAND.UN_SET,
        "data": app[OPERATION.UN_SET],
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb) {
        cb();
    };
};

CallRouter.prototype.__setUser = function (app, cb) {
    var prop = app[OPERATION.SET_USER];
    if (typeof prop['name'] !== 'string') {
        log.error('bad request setUser');
        if (cb) {
            cb();
        };
        return;
    };

    this.execApp({
        "app": FS_COMMAND.SET_USER,
        "data": prop['name'] + '@' + this.domain + (prop['prefix'] ? ' ' + prop['prefix'] : ''),
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb) {
        cb();
    };
};

CallRouter.prototype.__receiveFax = function (app, cb) {
    var prop = app[OPERATION.RECEIVE_FAX],
        _set = [],
        email = '';

    this.execApp({
        "app": FS_COMMAND.ANSWER
    });

    this.execApp({
        "app": FS_COMMAND.PLAYBACK,
        "data": "silence_stream://2000"
    });

    if (prop['enable_t38'] === true) {
        _set.push("fax_enable_t38_request=true", "fax_enable_t38=true");
    };

    if (prop['email'] instanceof Array) {
        email = prop['email'].join(',');
    };

    _set.push("execute_on_fax_success=lua FaxUpload.lua ${uuid} ${domain_name} " + email,
        "execute_on_fax_failure=system /bin/rm /recordings/${uuid}.tif");

    this.__setVar({
        "setVar": _set
    });

    this.execApp({
        "app": FS_COMMAND.RX_FAX,
        "data": "/recordings/${uuid}.tif"
    });

    if (cb)
        return cb();
};

CallRouter.prototype.__checkCallForward = function (app, cb) {
    var prop = app[OPERATION.CALL_FORWARD],
        status = this.getChnVar('Caller-Account-Status'),
        number = this.getChnVar('Caller-Account-Status-Description');
    log.trace('CF: status = %s', status);
    log.trace('CF: number = %s', number);
    if (status != 'CALLFORWARD') {
        if (cb)
            cb();
        return;
    };

    if (!number) {
        log.warn('bad request _callForward. SKIP application');
        if (cb) {
            cb();
        };
        return;
    };

    if (prop['user']) {
        this.execApp({
            "app": FS_COMMAND.SET_USER,
            "data": (prop.user['name'] || prop.user) + '@' + this.domain + (prop.user['prefix'] ? ' ' + prop.user['prefix'] : '')
        }, function (res) {
            console.dir(res);
        });
    };

    this.execApp({
        "app": FS_COMMAND.TRANSFER,
        "data": number + ' XML default'
    });

    this.end = true;
    if (cb) {
        cb();
    };
};

CallRouter.prototype.__blackList = function (app, cb) {
    var prop = app[OPERATION.BLACK_LIST],
        variableName = prop['variable'] || 'Channel-Caller-ID-Number',
        name = prop['name'] || '',
        scope = this,
        actions,
        number = this.getChnVar(variableName) || '';

    if (name == '' || number == '') {
        log.warn('Bad request __blackList');
        if (cb) {
            cb();
        };
        return;
    };

    number = number.replace(/\D/g, '');

    if (prop['actions'] instanceof Array && prop['actions'].length > 0) {
        actions = prop['actions'];
    } else {
        actions = [{
            "hangup": "INCOMING_CALL_BARRED"
        }];
    };
    actions.push({
        "break": true
    });
    blackList.check(this.domain, name, number, function (err, res) {
        if (err) {
            log.error(err);
            if (cb)
                cb();
            return;
        };

        if (res > 0) {
            log.trace('Black list number %s execute actions.', number);
            scope.execute(actions, cb);
        } else {
            log.trace('Black list skip number %s.', number);
            if (cb)
                cb();
        };
    });
};

CallRouter.prototype.__pickup = function (app, cb) {
    var groupName = app[OPERATION.PICKUP];
    if (typeof groupName !== 'string') {
        log.error('bad request __pickup');
        if (cb) {
            cb();
        };
        return;
    };

    this.execApp({
        "app": FS_COMMAND.PICKUP,
        "data": groupName + '@' + this.domain,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb) {
        cb();
    };
};

CallRouter.prototype.__ringback = function (app, cb) {
    let prop = app[OPERATION.RINGBACK];
    if (prop.hasOwnProperty('call') && prop.call['name']) {
        let call = prop.call;
        this.__setVar({
            "setVar": "ringback=" + this._getPlaybackFileString(call.type, call.name, call.refresh, true, call)
        });
    };

    if (prop.hasOwnProperty('hold')) {
        let hold = prop.hold;
        this.__setVar({
            "setVar": "hold_music=" + this._getPlaybackFileString(hold.type, hold.name, hold.refresh, true, hold)
        });
    };

    if (prop.hasOwnProperty('transfer') && prop.transfer['name']) {
        let transfer = prop.transfer;
        this.__setVar({
            "setVar": "transfer_ringback=" + this._getPlaybackFileString(transfer.type, transfer.name, transfer.refresh, true, transfer)
        });
    };

    if (cb)
        return cb();
};

CallRouter.prototype.__string = function (app, cb) {
    try {
        let prop = app[OPERATION.STRING] || {},
            data = this._parseVariable(prop.data),
            fn = prop.fn,
            varName = prop.setVar,
            args = prop.args
            ;

        if (!data || !varName) {
            log.error('Bad __string parameters');
            return cb && cb();
        }

        if (args instanceof Array) {
            // TODO
            for (let i = 0, len = args.length; i < len; i++) {
                if (typeof args[i] == 'string') {
                    var match = args[i].match(new RegExp('^/(.*?)/([gimy]*)$'));
                    if (match)
                        args[i] = new RegExp(match[1], match[2]);
                }
            }
        } else {
            args = [args];
        }

        let res;

        if (typeof StringOperation[fn] == 'function') {
            res = StringOperation[fn](data, args);
        } else if (typeof data[fn] == 'function') {
            res = data[fn].apply(data, args);
        } else {
            log.error('Bad __string fn name');
            return cb && cb();
        }
        this.__setVar({
            "setVar": varName + '=' + (res || '')
        }, cb);

    } catch (e) {
        log.error(e);
        return cb && cb();
    }
};

let StringOperation = {
    reverse: (s) => {
        if (!s)
            return '';

        return reverseString(s);
    }
};

CallRouter.prototype.__math = function (app, cb) {
    let prop = app[OPERATION.MATH],
        data = prop.data,
        varName = prop.setVar,
        fn = prop.fn || 'random',
        result
        ;
    if (!varName) {
        log.error('Bad varName');
        return cb && cb();
    };

    if (!data) {
        data = []
    } else if (typeof  data == 'string') {
        let _parseData = this._parseVariable(data);
        data = (_parseData && _parseData.split('')) || '';
    } else if ( !(data instanceof Array) ) {
        data = [data]
    }

    if (MathOperation.hasOwnProperty(fn)) {
        result = MathOperation[fn](data);
    } else if (typeof  Math[fn] == "function") {
        result = Math[fn].apply(null, data);
    } else if (Math.hasOwnProperty(fn)) {
        result = Math[fn];
    } else {
        log.error('Bad fn name: ', fn);
        return cb && cb();
    };

    this.__setVar({
        "setVar": varName + '=' + (result || '')
    }, cb);

};

var MathOperation = {
    'random': function (array) {
        let min = 0,
            max = array.length - 1
        ;
        return array[Math.floor(Math.random() * (max - min + 1) + min)]
    }
};


CallRouter.prototype.__eavesdrop = function (app, cb) {
    var prop = app[OPERATION.EAVESDROP],
        user = prop.user,
        spy = !!prop.spy
        ;

    if (!user || user.length < 1) {
        log.error('Bad eavesdrop parameters');
        return cb && cb();
    };

    this.__answer({
        "answer": ""
    });

    this.__setVar({
        "setVar": "webitel_direction=eavesdrop"
    });

    if (user === 'all') {
        this.__setVar({
            "setVar": "eavesdrop_require_group=" + this.domain
        });

        this.execApp({
            "app": FS_COMMAND.EAVESDROP,
            "data": this.domain,
            "async": app[OPERATION.ASYNC] ? true : false
        });
    } else {
        let data,
            number = this._parseVariable(user),
            fsApp = FS_COMMAND.EAVESDROP;
        if (spy) {
            fsApp = FS_COMMAND.USERSPY;
            data = number + '@${domain_name}';
        } else {
            data = '${hash(select/spymap/${domain_name}-' + number + ')}';
        }

        this.execApp({
            "app": fsApp,
            "data": data,
            "async": app[OPERATION.ASYNC] ? true : false
        });
    }

    return cb && cb();
};

CallRouter.prototype.__sipRedirect = function (app, cb) {
    var prop = app[OPERATION.SIP_REDIRECT],
        app  = FS_COMMAND.REDIRECT;
     if (+this.getChnVar('answer_epoch') > 0)
        app = FS_COMMAND.DEFLECT;

    this.execApp({
        "app": app,
        "data": '' + prop,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    return cb && cb();
};

CallRouter.prototype.__agent = function (app, cb) {
    var prop = app[OPERATION.AGENT],
        name = this._parseVariable(prop.name || '${caller_id_number}')
    ;

    if (!name) {
        log.error('Bad __agent options');
        return cb && cb();
    }
    name = name.replace(/@.*/, '') + `@${this.domain}`;
    let status = prop.status || "Available";
    let state = prop.state;// || "Waiting";

    this.api(`callcenter_config agent set status ${name} '${status}'`, (res) => log.debug(`${name} '${status}' => ${res.body}`));

    if (state)
        this.api(`callcenter_config agent set state ${name} '${state}'`, (res) => log.debug(`${name} '${state}' => ${res.body}`));

    return cb && cb();
};

CallRouter.prototype.api = function (str, cb) {
    log.trace('Exec %s', str);
    return this.connection.api(str, cb);
};

CallRouter.prototype.__avmd = function (app, cb) {
    let prop = app[OPERATION.AVMD] || {},
        _app = '',
        data = '';

    if (prop.action == "start") {
        let params = [];
        if (prop.simplifiedEstimation) {
            params.push(`simplified_estimation=${prop.simplifiedEstimation}`);
        }
        if (prop.inboundChannel) {
            params.push(`inbound_channel=${prop.inboundChannel}`);
        }
        if (prop.outboundChannel) {
            params.push(`outbound_channel=${prop.outboundChannel}`);
        }
        if (prop.continuousStreak) {
            params.push(`sample_n_continuous_streak=${prop.continuousStreak}`);
        }
        if (prop.toSkip) {
            params.push(`sample_n_to_skip=${prop.toSkip}`);
        }
        if (prop.debug) {
            params.push(`debug=${prop.debug}`);
        }
        if (prop.reportStatus) {
            params.push(`report_status=${prop.reportStatus}`);
        }

        data = params.join(',');
        _app = FS_COMMAND.AVMD_START;
    } else if (prop.action == "stop") {
        _app = FS_COMMAND.AVMD_STOP;
    } else {
        log.error(`Bad __avmd action parameters.`);
        return cb && cb();
    }

    this.execApp({
        "app": _app,
        "data": data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    return cb && cb();
};