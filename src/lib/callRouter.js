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

function push(arr, e) {
    arr.push(e);
    return arr.length - 1;
};

function equalsRange (_curentDay, _tmp) {
    var _min, _max;

    _tmp = _tmp.split('-');
    _min = parseInt(_tmp[0]);
    _max = _tmp[1]
        ? parseInt(_tmp[1])
        : 7;
    if (_min > _max) {
        _tmp = _max;
        _max = _min;
        _min = _tmp;
    };
    return (_curentDay >= _min && _curentDay <= _max);
};

var CallRouter = module.exports = function (connection, globalVar, regCollection, timeOffset) {
    this.globalVar = globalVar || {};
    this.connection = connection;
    this.regCollection = regCollection || {};
    this.timeOffset = timeOffset || 0;
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

CallRouter.prototype.wday = function (param) {
    this._DateParser(param, (new Date().getDay() + 1));
};

CallRouter.prototype._DateParser = function (param, datetime) {
    param = param || '';
    var datetimes = param.replace(/\s/g, '').split(','),
        result = false;
    if (datetimes[0] == "") {
        throw Error("&wday bad parameters");
    };
    for (var i = 0; i < datetimes.length; i++) {
        result = (datetimes[i].indexOf('-') == -1)
            ? datetime == parseInt(datetimes[i])
            : equalsRange(datetime, datetimes[i]);

        if (result == true) {
            return result
        };
    };
    return result;
};

CallRouter.prototype.year = function (param) {
    param = param || '';
    var days = param.replace(/\s/g, '').split(','),
        _curentDay = new Date().getYear(),
        result = false;
    if (days[0] == "") {
        throw Error("&wday bad parameters");
    };
    for (var i = 0; i < days.length; i++) {
        result = (days[i].indexOf('-') == -1)
            ? _curentDay == parseInt(days[i])
            : equalsRange(_curentDay, days[i]);

        if (result == true) {
            return result
        };
    };
    return result;
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