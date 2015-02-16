/**
 * Created by i.navrotskyj on 24.01.2015.
 */

var log = require('./log')(module),
    vm = require('vm');

var OPERATION = {
    IF: "if",
    THEN: "then",
    ELSE: "else",
    APPLICATION: "app",
    DATA: "data",
    ASYNC: "async",

    ECHO: "echo",

    ANSWER: "answer",
    SET: "set",
    GOTO: "goto",
    GATEWAY: "gateway",
    DEVICE: "device",
    RECORD_SESSION: "record_session",
    HANGUP: "hangup",
    SCRIPT: "script",
    LOG: "log"
};

var FS_COMMAND = {
    ANSWER: "answer",
    PRE_ANSWER: "pre_answer",
    RING_READY: "ring_ready",
    TRANSFER: "transfer",
    HANGUP: "hangup",
    BRIDGE: "bridge",

    SET: "set",
    MULTISET: "multiset",
    EXPORT: "export",

    RECORD_SESSION: "record_session",
    STOP_RECORD_SESSION: "stop_record_session",

    LUA: "lua",
    JS: "js",

    LOG: "log",

    ECHO: "echo",
    DELAY_ECHO: "delay_echo"
};


var COMMANDS = {
    REGEXP: "&reg"
};

var CallRouter = module.exports = function (connection, globalVar, desNumber, chnNumber, timeOffset, versionSchema) {
    this.globalVar = globalVar || {};
    this.connection = connection;
    this.regCollection = {};
    this.offset = timeOffset;
    this.versionSchema = versionSchema;
    this.setDestinationNumber(desNumber, chnNumber);
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
    return new Date(utc + (3600000 * this.offset));
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
    _reg = reg.match(new RegExp('^/(.*?)/([gimy]*)$'));
    if (!_reg) {
        _reg = [null, reg];
    };
    var _result = new RegExp(_reg[1], _reg[2]).exec(val);
    var _regOb = {};
    // TODO оптимизировать
    for (var key in _result) {
        _regOb['$' + key] = _result[key];
    };
    this.regCollection[COMMANDS.REGEXP + (Object.keys(this.regCollection).length + 1)] = _regOb;
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

CallRouter.prototype.execIf = function (condition) {
    var sandbox = {
        _resultCondition: false,
        sys: this
    };
    if (condition['sysExpression']) {
        var expression = condition['sysExpression'] || '';

        log.info('Parse expression: %s', expression);
            // TODO
       var script = vm.createScript('_resultCondition = (' + expression + ')');
       script.runInNewContext(sandbox);

        log.trace('Condition %s : %s', expression, sandbox._resultCondition
            ? true
            : false);
        if (sandbox._resultCondition) {
            if (condition[OPERATION.THEN]) {
                this.execute(condition[OPERATION.THEN]);
            };
        } else {
            if (condition[OPERATION.ELSE]) {
                this.execute(condition[OPERATION.ELSE]);
            };
        };
    };
};

CallRouter.prototype.execApp = function (_obj) {
    if (_obj[OPERATION.APPLICATION]) {
        if (_obj[OPERATION.DATA]) {
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
        this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA] || '');
    };
};

CallRouter.prototype.doExec = function (condition) {
    if (condition instanceof Object) {

        if (this.versionSchema === 2) {

            if (condition.hasOwnProperty(OPERATION.IF)) {
                this.execIf(condition[OPERATION.IF]);
            }
            else if (condition.hasOwnProperty(OPERATION.ANSWER)) {
                this._answer(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.SET)) {
                this._set(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.GOTO)) {
                this._goto(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.GATEWAY)) {
                this._gateway(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.DEVICE)) {
                this._device(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.RECORD_SESSION)) {
                this._recordSession(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.HANGUP)) {
                this._hangup(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.SCRIPT)) {
                this._script(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.LOG)) {
                this._console(condition);
            }
            else if (condition.hasOwnProperty(OPERATION.ECHO)) {
                this._echo(condition);
            }
            else {
                log.error('error parse json');
            };
        } else {
            if (condition.hasOwnProperty(OPERATION.IF)) {
                this.execIf(condition[OPERATION.IF]);
            } else if (condition.hasOwnProperty(OPERATION.APPLICATION)) {
                this.execApp(condition);
            } else {
                log.error('error parse json');
            };
        }

    };
};

CallRouter.prototype.execute = function (callflows) {
    var scope = this;
    if (callflows instanceof Array && callflows.length > 0) {
        callflows.forEach(function (callflow) {
            scope.doExec(callflow);
        });
    };
};

CallRouter.prototype.start = function (callflows) {
    var scope = this;
    if (callflows instanceof Array && callflows.length > 0) {
        callflows.forEach(function (callflow) {
            scope.destroyLocalRegExpValues();
            scope.execute([callflow]);
        });
    };
};

CallRouter.prototype._answer = function (app) {
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
            "app": _app
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.ANSWER]);
    };
};

CallRouter.prototype._set = function (app) {
    var _app, _data;
    if (app[OPERATION.SET].indexOf('all:') == 0) {
        _app = FS_COMMAND.EXPORT;
        _data = app[OPERATION.SET].substring(4);
    } else if (app[OPERATION.SET].indexOf('nolocal:') == 0) {
        _app = FS_COMMAND.EXPORT;
        _data = app[OPERATION.SET];
    } else if (/(\w|{|}|&|&|\$|\$\$|-|\s|'|")*=(\w|{|}|&|\$|\$\$|-|\s|'|")*,/.test(app[OPERATION.SET])) {
        _app = FS_COMMAND.MULTISET;
        _data = '^^,' + app[OPERATION.SET]
    } else {
        _app = FS_COMMAND.SET;
        _data = app[OPERATION.SET];
    };

    if (_app) {
        this.execApp({
            "app": _app,
            "data": _data
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.SET]);
    };
};

CallRouter.prototype._goto = function (app) {
    var _app = FS_COMMAND.TRANSFER,
        _data;
    if (app[OPERATION.GOTO].indexOf('default:') == 0) {
        _data = app[OPERATION.GOTO].substring(8) + ' XML default';
    } else if (app[OPERATION.GOTO].indexOf('public:') == 0) {
        _data = app[OPERATION.GOTO].substring(7) + ' XML public';
    } else {
        _data = app[OPERATION.GOTO];
    };

    this.execApp({
        "app": _app,
        "data": _data
    });
};

CallRouter.prototype._gateway = function (app) {
    var _data;
    if (app[OPERATION.GATEWAY].indexOf('sip:') == 0) {
        _data = 'sofia/gateway/' + app[OPERATION.GATEWAY].substring(4);
    };

    if (_data) {
        this.execApp({
            "app": FS_COMMAND.BRIDGE,
            "data": _data
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.GATEWAY]);
    };
};

CallRouter.prototype._device = function (app) {
    if (typeof app[OPERATION.DEVICE] == 'string' && app[OPERATION.DEVICE] != '') {
        this.execApp({
            "app": FS_COMMAND.BRIDGE,
            "data": "user/" + app[OPERATION.DEVICE] + "@${domain_name}"
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.DEVICE]);
    };
};

CallRouter.prototype._recordSession = function (app) {
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
};

CallRouter.prototype._hangup = function (app) {
    this.execApp({
        "app": FS_COMMAND.HANGUP,
        "data": app[OPERATION.HANGUP] || ''
    });
};

CallRouter.prototype._script = function (app) {
    var _data,
        _app;
    if (app[OPERATION.SCRIPT].indexOf('lua:') == 0) {
        _app = FS_COMMAND.LUA;
        _data = 'lua/' + app[OPERATION.SCRIPT].substring(4);
    } else if (app[OPERATION.SCRIPT].indexOf('js:') == 0) {
        _app = FS_COMMAND.JS;
        _data = 'js/' + app[OPERATION.SCRIPT].substring(3);
    } else {
        log.warn('Bad parameter ', app[OPERATION.SCRIPT]);
        return false;
    };

    this.execApp({
        "app": _app,
        "data": _data
    });
};

CallRouter.prototype._console = function (app) {
    if (typeof app[OPERATION.LOG] == 'string') {
        this.execApp({
            "app": FS_COMMAND.LOG,
            "data": 'CONSOLE ' + app[OPERATION.LOG]
        });
    } else {
        log.warn('Bad parameter ', app[OPERATION.SCRIPT]);
        return false;
    };
};

CallRouter.prototype._echo = function (app) {
    var _app, _data = '', delay = parseInt(app[OPERATION.ECHO]);
    if (delay > 0) {
        _app = FS_COMMAND.DELAY_ECHO;
        _data = app[OPERATION.ECHO];
    } else {
        _app = FS_COMMAND.ECHO;
    };
    this.execApp({
        "app": _app,
        "data": _data
    });
};