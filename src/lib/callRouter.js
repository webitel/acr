/**
 * Created by i.navrotskyj on 24.01.2015.
 * http://stackoverflow.com/questions/20373746/parsing-operators-and-evaluating-them-in-javascript
 * http://jsep.from.so/
 */
var log = require('./log')(module);

var OPERATION = {
    IF: "if",
    THEN: "then",
    ELSE: "else",
    APPLICATION: "app",
    DATA: "data"
};

var build = module.exports = function (connection) {
    this.connection = connection;
};

build.prototype.read = function (expression) {
    var variables = expression.match(/\$\w+/g);
    var length = variables
        ? variables.length
        : 0;
    var uniqueVariables = [];
    var index = 0;

    while (index < length) {
        var variable = variables[index++];
        if (uniqueVariables.indexOf(variable) < 0)
            uniqueVariables.push(variable);
    }

    return Function.apply(null, uniqueVariables.concat("return " + expression));
};

build.prototype.execIf = function (condition) {
    if (condition['expression']) {
        var expression = this.read(condition['expression']);
        log.trace('Condition %s = %s', condition['expression'], expression(''));
        if (expression('')) {
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

build.prototype.execApp = function (_obj) {
    if (_obj[OPERATION.APPLICATION]) {
        log.trace('Execute app: %s, with data: %s', _obj[OPERATION.APPLICATION], _obj[OPERATION.DATA]);
        this.connection.execute(_obj[OPERATION.APPLICATION], _obj[OPERATION.DATA])
    }
};

build.prototype.doExec = function (extension) {
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