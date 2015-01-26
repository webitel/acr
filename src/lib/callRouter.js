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
    ASYNC: "async"
};

var CallRouter = module.exports = function (connection) {
    this.connection = connection;
};

CallRouter.prototype.read = function (expression) {
    return expression.replace(/\$\{([\s\S]*?)\}/g, function (v) {
        return '_chnData.getHeader("' + v.substring(2, v.length - 1) + '")'
    });
};

CallRouter.prototype.execIf = function (condition) {
    var sandbox = {
        _resultCondition: false,
        _chnData: this.connection.channelData
    };
    if (condition['expression']) {
        var expression = this.read(condition['expression']);

        try {
            console.log('_resultCondition = (' + expression + ')');
            var script = vm.createScript('_resultCondition = (' + expression + ')');
            script.runInNewContext(sandbox);
        } catch (e) {
            log.error(e.message);
        }
        log.trace('Condition %s : %s', condition['expression'], sandbox._resultCondition
            ? true
            : false);
        if (sandbox._resultCondition) {
            if (condition[OPERATION.THEN]) {
                this.doExec(condition[OPERATION.THEN])
            }
        } else {
            if (condition[OPERATION.ELSE]) {
                this.doExec(condition[OPERATION.ELSE])
            }
        }
    }
};

CallRouter.prototype.execApp = function (_obj) {
    if (_obj[OPERATION.APPLICATION]) {
        if (_obj[OPERATION.ASYNC]) {
            log.trace('Execute sync app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
            this.connection.setEventLock(false);
            this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
        } else {
            this.connection.setEventLock(true);
            log.trace('Execute app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
            this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
        }
    }
};

CallRouter.prototype.doExec = function (callflow) {
    var condition;

    if (callflow instanceof Array && callflow.length > 0) {
        for (var key in callflow) {
            condition = callflow[key];
            if (condition instanceof Object && condition.hasOwnProperty(OPERATION.IF)) {
                this.execIf(condition[OPERATION.IF]);
            } else if (condition instanceof Object && condition.hasOwnProperty(OPERATION.APPLICATION)) {
                this.execApp(condition);
            } else {
                log.error('error parse json');
            }
        }
    }
};