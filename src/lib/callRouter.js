/**
 * Created by i.navrotskyj on 24.01.2015.
 * http://stackoverflow.com/questions/20373746/parsing-operators-and-evaluating-them-in-javascript
 * http://jsep.from.so/
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

var keywords = 'function|case|if|return|new|switch|var|this|typeof|for|while|break|do|continue';

var CallRouter = module.exports = function (connection, globalVar, regCollection) {
    this.globalVar = globalVar || {};
    this.connection = connection;
    this.regCollection = regCollection || {};
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
    return '\'' + _var + '\'';
};


CallRouter.prototype.getGlbVar = function (name) {
    try {
        var _var = this.globalVar[0][name];
        return _var
            ? ('\'' + _var + '\'')
            : ''
    } catch (e) {
        return '';
    }
};

CallRouter.prototype.execIf = function (condition) {
    var scope = this;
    var sandbox = {
        _resultCondition: false
    };
    if (condition['expression']) {

        function push(arr, e) {
            arr.push(e);
            return arr.length - 1;
        };
        var all = [];

        var expression = condition['expression'] || '';
        expression = expression
            // GLOBAL
            .replace(/\$\$\{([\s\S]*?)\}/gi, function (a, b) {
                return scope.getGlbVar(b);
            })
            // ChannelVar
            .replace(/\$\{([\s\S]*?)\}/gi, function (a, b) {
                return scope.getChnVar(b);
            })
            .replace(/(\/(\\\/|[^\/\n])*\/[gim]{0,3})|(([^\\])((?:'(?:\\'|[^'])*')|(?:"(?:\\"|[^"])*")))/g, function(m, r, d1, d2, f, s, b, bb)
            {
                if (r != null && r != '') {
                    s = r;
                    m = '\0B';

                } else {
                    s = s;
                    m = f + '\0B';
                }
                return m + push(all, s) + '\0';
            })
            .replace(new RegExp('\\b(' + keywords + ')\\b', 'gi'), '')
            .replace(/\&match\(([\s\S]*?)\)/gi, function (f, param) {
                var _params = param.split(',', 2),
                    _reg, _val;

                _reg = (/\0B(\d+)\0/g.test(_params[0]))
                   ? all[_params[0].replace(/\D/g, '')].match(new RegExp('^/(.*?)/([gimy]*)$'))
                   : _params[0].match(new RegExp('^/(.*?)/([gimy]*)$'));

                _val = (/\0B(\d+)\0/g.test(_params[1]))
                    ? all[_params[1].replace(/\D/g, '')].replace(/\'/g, '')
                    : _params[1];

                var _result = new RegExp(_reg[1], _reg[2]).exec(_val);
                var _regOb = {};
                for (var key in _result) {
                    _regOb['$' + key] = _result[key];
                };
                scope.regCollection[COMMANDS.REGEXP + (Object.keys(scope.regCollection).length + 1)] = _regOb;
                return _result ? true : false;
            })
            .replace(/\0B(\d+)\0/g, function(m, i) {
                return all[i];
            });

        try {
            // TODO
           var script = vm.createScript('_resultCondition = (' + expression + ')');
           script.runInNewContext(sandbox);

        } catch (e) {
            log.error(e.message);
        };
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
    }
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