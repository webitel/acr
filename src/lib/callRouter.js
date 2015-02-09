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
    CALL_RECORD: "callRecord"
};

var COMMANDS = {
    REGEXP: "&reg"
};

var CallRouter = module.exports = function (connection, globalVar, desNumber, chnNumber, timeOffset) {
    this.globalVar = globalVar || {};
    this.connection = connection;
    this.regCollection = {};
    this.timeOffset = timeOffset || 0;
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
    return this._DateParser(param || '', (new Date().getFullYear()), 9999);
};

CallRouter.prototype.yday = function (param) {
    var now = new Date(),
        start = new Date(now.getFullYear(), 0, 0),
        diff = now - start,
        oneDay = 1000*60*60*24;
    return this._DateParser(param, (Math.floor(diff / oneDay)), 366);
};

CallRouter.prototype.mon = function (param) {
    return this._DateParser(param, (new Date().getMonth() + 1), 12);
};

CallRouter.prototype.mday = function (param) {
    return this._DateParser(param, new Date().getDate(), 31);
};

CallRouter.prototype.week = function (param) {
    return this._DateParser(param, new Date()._getWeek(), 53);
};

CallRouter.prototype.mweek = function (param) {
    return this._DateParser(param, (new Date()._getWeekOfMonth() + 1), 6);
};

CallRouter.prototype.wday = function (param) {
    return this._DateParser(param, (new Date().getDay() + 1), 7);
};

CallRouter.prototype.hour = function (param) {
    return this._DateParser(param, new Date().getHours(), 23);
};

CallRouter.prototype.minute = function (param) {
    return this._DateParser(param, new Date().getMinutes(), 59);
};

CallRouter.prototype.minute_of_day = function (param) {
    var now = new Date();
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
    if ( !(strReg instanceof String) ) return;
    var _tmp = strReg.match(new RegExp('^/(.*?)/([gimy]*)$')),
        _reg = new RegExp(_tmp[1], _tmp[2]).exec(number);
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

CallRouter.prototype.execCallRecord = function (_obj) {
    var _quality = _obj[OPERATION.CALL_RECORD] || '16';
    this.execApp({
        'app': 'lua',
        'data': 'RecordSession.lua ' + _quality
    });
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
            log.trace('Execute sync app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
            this.connection.setEventLock(false);
        } else {
            this.connection.setEventLock(true);
            log.trace('Execute app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
        };
        this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
    };
};

CallRouter.prototype.doExec = function (condition) {
    if (condition instanceof Object && condition.hasOwnProperty(OPERATION.IF)) {
        this.execIf(condition[OPERATION.IF]);
    } else if (condition instanceof Object && condition.hasOwnProperty(OPERATION.APPLICATION)) {
        this.execApp(condition);
    } else if (condition instanceof Object && condition.hasOwnProperty(OPERATION.CALL_RECORD)) {
        this.execCallRecord(condition);
    } else {
        log.error('error parse json');
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