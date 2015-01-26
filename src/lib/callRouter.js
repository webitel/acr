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
    DATA: "data"
};

var CallRouter = module.exports = function (connection) {
    this.connection = connection;
};

CallRouter.prototype.read = function (callflow) {
    return callflow.replace(/\$\{([\s\S]*?)\}/g, function (v) {
        return '_chnData.getHeader("' + v.substring(2, v.length - 1) + '")'
    });
};

CallRouter.prototype.execIf = function (condition) {
    var sandbox = {
        _resultCondition: false,
        _chnData: this.connection.channelData
    };
    if (condition['callflow']) {
        var callflow = this.read(condition['callflow']);

        try {
            var script = vm.createScript('_resultCondition = (' + callflow + ')');
            script.runInNewContext(sandbox);
        } catch (e) {
            log.error(e.message);
        }
        log.trace('Condition %s : %s', condition['callflow'], sandbox._resultCondition
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
        log.trace('Execute app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
        this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA])
    }
};

CallRouter.prototype.doExec = function (extension) {
    var condition;

    if (extension instanceof Array && extension.length > 0) {
        for (var key in extension) {
            condition = extension[key];
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