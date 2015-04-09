/**
 * Created by i.navrotskyj on 24.01.2015.
 */

var log = require('./log')(module),
    httpReq = require('./httpRequest'),
    calendar = require('../middleware/calendar');

var MEDIA_TYPE = {
    WAV: 'wav',
    MP3: 'mp3',
    LOCAL: 'local'
};

var OPERATION = {
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
    QUEUE: "queue"
};

var FS_COMMAND = {
    ANSWER: "answer",
    PRE_ANSWER: "pre_answer",
    RING_READY: "ring_ready",
    TRANSFER: "transfer",
    HANGUP: "hangup",

    SET: "set",
    MULTISET: "multiset",
    EXPORT: "export",

    RECORD_SESSION: "record_session",
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
    PLAY_AND_GET: "play_and_get_digits",
    PARK: "park",
    CALLCENTER: "callcenter"
};


var COMMANDS = {
    REGEXP: "&reg"
};

var MAX_CYCLE_COUNT = 20;

var CallRouter = module.exports = function (connection, option) {
    option = option || {};
    this.index = 0;
    this.cycleCount = 0;
    this.globalVar = option['globalVar'] || {};
    this.connection = connection;
    this.regCollection = {};
    this.offset = option['timeOffset'];
    this.domain = option['domain'];
    this.versionSchema = option['versionSchema'];
    this.setDestinationNumber(option['desNumber'], option['chnNumber']);
};

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


CallRouter.prototype.destroyLocalRegExpValues = function () {
    var scope = this;
    Object.keys(this.regCollection).forEach(function (key) {
        if (key != (COMMANDS.REGEXP + 0)) {
            delete scope.regCollection[key];
        }
    });
};

CallRouter.prototype.DateOffset = function() {
    if (!this.offset) {
        return new Date();
    };
    var d = new Date(),
        utc = d.getTime() + (d.getTimezoneOffset() * 60000);
    return new Date(utc + (60000 * Math.abs(this.offset) ));
};

CallRouter.prototype.setChnVar = function (name, value) {
    this.connection.channelData.addHeader(name, value);
};

CallRouter.prototype.getChnVar = function (name) {
    var _var = this.connection.channelData.getHeader('variable_' + name)
        || this.connection.channelData.getHeader(name)
        || '';
    return _var ;
};

CallRouter.prototype.getGlbVar = function (name) {
    try {
        var _var = this.globalVar[0][name];
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

CallRouter.prototype.time_of_day = function (param) {
    // TODO
};

CallRouter.prototype._DateParser = function (param, datetime, maxVal) {
    param = param || '';
    var datetimes = param.replace(/\s/g, '').split(','),
        result = false;
    if (datetimes[0] == "") {
        throw Error("bad parameters");
    };
    for (var i = 0; i < datetimes.length; i++) {
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

CallRouter.prototype._switch = function (condition, cb) {
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

CallRouter.prototype.execIf = function (condition, cb) {
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
            log.error(e.message);
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
    };
};

CallRouter.prototype._parseVariable = function (name) {
    var scope = this;
    name = name || '';
    try {
        return name
            .replace(/\$\$\{([\s\S]*?)\}/gi, function (a, b) {
                d
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

CallRouter.prototype.doExec = function (condition, cb) {
    if (condition instanceof Object) {

        if (this.versionSchema === 2) {

            if (condition.hasOwnProperty(OPERATION.IF)) {
                this.execIf(condition[OPERATION.IF], cb);
            } else if (condition.hasOwnProperty(OPERATION.SWITCH)) {
                this._switch(condition[OPERATION.SWITCH], cb);
            }
            else if (condition.hasOwnProperty(OPERATION.SLEEP)) {
                this._sleep(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.ANSWER)) {
                this._answer(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.BRIDGE)) {
                this._bridge(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.SET)) {
                this._set(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.GOTO)) {
                this._goto(condition, cb);
            } /*
            else if (condition.hasOwnProperty(OPERATION.GATEWAY)) {
                this._gateway(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.DEVICE)) {
                this._device(condition, cb);
            }*/
            else if (condition.hasOwnProperty(OPERATION.RECORD_SESSION)) {
                this._recordSession(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.HANGUP)) {
                this._hangup(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.SCRIPT)) {
                this._script(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.LOG)) {
                this._console(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.ECHO)) {
                this._echo(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.HTTP)) {
                this._httpRequest(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.CONFERENCE)) {
                this._conference(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.SCHEDULE)) {
                this._schedule(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.PLAYBACK)) {
                this._playback(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.BREAK)) {
                this._break(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.CALENDAR)) {
                this._calendar(condition, cb);
            }
            else if (condition.hasOwnProperty(OPERATION.PARK)) {
                this._park(condition, cb);
            } else if (condition.hasOwnProperty(OPERATION.QUEUE)) {
                this._queue(condition, cb);
            }
            else {
                log.error('error parse json');
            };
        } else {
            if (condition.hasOwnProperty(OPERATION.IF)) {
                this.execIf(condition[OPERATION.IF], cb);
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
        }

    };
};

CallRouter.prototype._calendar = function (condition, cb) {
    var exportVariable = condition[OPERATION.CALENDAR]['exportVariable'],
        scope = this
        ;
    if (!exportVariable) {
        log.trace('_calendar: Bad parameters.');
        if (cb) {
            cb(new Error('Bad parameters.'))
        };
        return;
    };
    calendar(this, condition[OPERATION.CALENDAR], function (err, res) {
        try {
            if (err) {
                log.error(err['message']);
                if (cb)
                    cb(err);
                return;
            };
            scope._set({
                "setVar": ''.concat(exportVariable, '=', res.length)
            }, cb);
        } catch (e) {
            log.error(e['message']);
            if (cb)
                cb(e);
        };
    });
};

CallRouter.prototype.execute = function (callflows, cb) {
    var scope = this;
    var i = 0;

    var postExec = function (err, res) {
        i++;
        if (i == callflows.length) {
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

CallRouter.prototype.start = function (callflows) {
    this.callflows = callflows;
    var scope = this;

    var postExec = function (err, res) {
        scope.index++;
        if (scope.index == callflows.length) {
            scope.connection.disconnect();
            return;
        };
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
    };
};

CallRouter.prototype._answer = function (app, cb) {
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

CallRouter.prototype._park = function (app, cb) {
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
        var scope = this, _tmp;
        variables.forEach(function(variableStr) {
            if (typeof variableStr != 'string') return;
            _tmp = variableStr.split('=');
            scope.setChnVar('variable_' + _tmp[0], _tmp[1]);
        });
    };
};

CallRouter.prototype._set = function (app, cb) {
    var _app, _data, _chnArrayVar;

    if (app[OPERATION.SET] instanceof Array) {
        _app = FS_COMMAND.MULTISET;
        _data = '^^~' + app[OPERATION.SET].join('~');
        _chnArrayVar = app[OPERATION.SET];
    } else {
        if (app[OPERATION.SET].indexOf('all:') == 0) {
            _app = FS_COMMAND.EXPORT;
            _data = app[OPERATION.SET].substring(4);
            _chnArrayVar = [_data];
        } else if (app[OPERATION.SET].indexOf('nolocal:') == 0) {
            _app = FS_COMMAND.EXPORT;
            _data = app[OPERATION.SET];
        } else {
            _app = FS_COMMAND.SET;
            _data = app[OPERATION.SET];
            _chnArrayVar = [_data];
        };
    };

    if (_app) {
        this.execApp({
            "app": _app,
            "data": _data,
            "async": app[OPERATION.ASYNC] ? true : false
        });
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


CallRouter.prototype._break = function (app, cb) {
    cb(new Error('BREAK'));
};

CallRouter.prototype._goto = function (app, cb) {
    var _app = FS_COMMAND.TRANSFER;
    if (app[OPERATION.GOTO] && app[OPERATION.GOTO].indexOf('local:') === 0) {
        var _i = parseInt(app[OPERATION.GOTO].substring(6));
        if (!isNaN(_i) && this.index !== _i) {
            log.trace('GOTO ' + _i);
            this.index = _i;
            this.cycleCount++;
            if (this.cycleCount === MAX_CYCLE_COUNT) {
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

CallRouter.prototype._recordSession = function (app, cb) {
    if (app[OPERATION.RECORD_SESSION] == 'start' || app[OPERATION.RECORD_SESSION] == '') {
        this.execApp({
            "app": "multiset",
            "data": "^^,RECORD_MIN_SEC=2,RECORD_STEREO=true,RECORD_BRIDGE_REQ=true," +
                "record_post_process_exec_api=luarun:RecordUpload.lua ${uuid} ${domain_name}"
        });

        this.execApp({
            "app": FS_COMMAND.RECORD_SESSION,
            "data": "/recordings/${uuid}.mp3"
        });
    } else if (app[OPERATION.RECORD_SESSION] == 'stop') {
        this.execApp({
            "app": FS_COMMAND.STOP_RECORD_SESSION,
            "data": "/recordings/${uuid}.mp3"
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.RECORD_SESSION]);
    };

    if (cb)
        cb();
};

CallRouter.prototype._hangup = function (app, cb) {
    this.execApp({
        "app": FS_COMMAND.HANGUP,
        "data": app[OPERATION.HANGUP] || '',
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};

CallRouter.prototype._script = function (app, cb) {
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

CallRouter.prototype._console = function (app, cb) {
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

CallRouter.prototype._echo = function (app, cb) {
    var _app, _data = '', delay = parseInt(app[OPERATION.ECHO]);
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

CallRouter.prototype._httpRequest = function (app, cb) {
    httpReq(app[OPERATION.HTTP], this, cb);
};

CallRouter.prototype._sleep = function (app, cb) {
    var delay = parseInt(app[OPERATION.SLEEP]);
    this.execApp({
        "app": FS_COMMAND.SLEEP,
        "data": delay,
        "async": app[OPERATION.ASYNC] ? true : false
    });
    if (cb)
        cb();
};

CallRouter.prototype._conference = function (app, cb) {
    var _data = '', prop = app[OPERATION.CONFERENCE];
    if (prop['name'] && /^[a-zA-Z0-9+_-]+$/.test(prop['name'])) {
        _data = _data.concat(prop['name'], '@',
            prop.hasOwnProperty('profile') ? prop['profile'] : 'default',
            prop.hasOwnProperty('pin') ? '+' + prop['pin'] : '',
            prop.hasOwnProperty('mute') ? '+flags{mute}' : ''
        );
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

CallRouter.prototype._schedule = function (app, cb) {
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

CallRouter.prototype._playback = function (app, cb) {
    var _fileName = app[OPERATION.PLAYBACK]["name"],
        type = app[OPERATION.PLAYBACK]["type"],
        filePath = '',
        scope = this;

    if (typeof _fileName !== 'string') {
        log.warn('Bad _playback parameters');
        if (cb)
            cb();
        return;
    };

    switch (type) {
        case MEDIA_TYPE.WAV:
            // TODO
            break;
        case MEDIA_TYPE.LOCAL:
            filePath = _fileName;
            break;
        default :
            var cdrUrl = this.getGlbVar('cdr_url');
            if (cdrUrl) {
                filePath = encodeURI(cdrUrl.replace(/https?/, 'shout') + '/sys/media/' + MEDIA_TYPE.MP3 + '/' + _fileName
                    + '?domain=' + this.domain);
            };
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
            "data": [_min, _max, _tries, _timeout, '#', filePath, 'silence_stream://250', _setVar, '\\d+'].join(' '),
            "async": app[OPERATION.PLAYBACK][OPERATION.ASYNC] ? true : false
        }, function (res) {
            try {
                var _r = res.getHeader('variable_' + _setVar) || '';
                scope._addVariableArrayToChannelDump([_setVar + '=' + _r]);
                log.trace('Set %s = %s', _setVar, _r);
            } catch (e) {
                log.error(e.message);
            };
            if (cb)
                cb();
        });

    } else {
        this.execApp({
            "app": FS_COMMAND.PLAYBACK,
            "data": filePath,
            "async": app[OPERATION.PLAYBACK][OPERATION.ASYNC] ? true : false
        });
        if (cb)
            cb();
    };
};

CallRouter.prototype._bridge = function (app, cb) {
    var prop = app[OPERATION.BRIDGE],
        _data = '',
        separator = prop['strategy'] == 'failover' // TODO переделать
            ? '|'
            : ','; // ":_:" - only for user & device; "," - for other types

    if (prop.hasOwnProperty('parameters') && prop['parameters'] instanceof Array) {
        _data = _data.concat('{', prop['parameters'].join(','), '}');
    };

    if (prop.hasOwnProperty('endpoints') && prop['endpoints'] instanceof Array) {
        prop['endpoints'].forEach(function (endpoint) {
            if (endpoint.hasOwnProperty('parameters') && endpoint['parameters'] instanceof Array) {
                _data = _data.concat('[', endpoint['parameters'].join(','), ']')
            };
            switch (endpoint['type']) {
                case 'sipGateway':
                    _data = _data.concat('sofia/gateway/', endpoint['name'], '/', endpoint['dialString']);
                    break;
                case 'sipUri':
                    _data = _data.concat('sofia/', endpoint.hasOwnProperty('profile') ? endpoint['profile'] : 'external',
                        '/', endpoint['dialString'], '@', endpoint['host']);
                    break;
                case 'sipDevice':
                    _data = _data.concat('sofia/', endpoint.hasOwnProperty('profile') ? endpoint['profile'] : 'external',
                        '/', endpoint['name'], '%', endpoint['domainName'], '^', endpoint['dialString']);
                    break;
                case 'device':
                    _data = _data.concat('user/', endpoint['name'], '@${domain_name}');
                    break;
                case 'user':
                    _data = _data.concat('user/', endpoint['name'], '@', endpoint.hasOwnProperty('domainName')
                        ? endpoint['domainName']
                        : '${domain_name}');
                    break;
            };
            _data = _data.concat(separator);
        });

        this.execApp({
            "app": FS_COMMAND.BRIDGE,
            "data": _data.slice(0, (-1 * separator.length)),
            "async": prop[OPERATION.ASYNC] ? true : false
        });
    };

    if (cb)
        cb();
};

CallRouter.prototype._queue = function (app, cb) {
    var _data = '', prop = app[OPERATION.QUEUE];
    if (prop['name'] && /^[a-zA-Z0-9+_-]+$/.test(prop['name'])) {
        _data = prop['name'] + '@${domain_name}';
    } else {
        log.error('Bad parameters queue.');
    };

    this.execApp({
        "app": FS_COMMAND.CALLCENTER,
        "data": _data,
        "async": app[OPERATION.ASYNC] ? true : false
    });

    if (cb)
        cb();
};