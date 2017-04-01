/**
 * Created by igor on 29.03.17.
 */

"use strict";

const BaseNode = require('./baseNode'),
    log = require(__appRoot + '/lib/log')(module),
    moment = require('moment-timezone'),
    Node = require('./node');

class ConditionNode extends BaseNode {
    constructor (parent, args = {}, options = {}) {
        super(parent, options);
        
        this.expression = args.sysExpression;

        this.then = new Node(parent);
        this.else = new Node(parent);
    }

    getThenNode () {
        this.then.first();
        return this.then;
    }

    getElseNode () {
        this.then.first();
        return this.else;
    }

    execute (call, cb) {

        if (!this.expression)
            return cb(new Error(`Bad application expression`));

        const result = execExpression(this.expression, getFn(call));

        call.log(`Expression ${this.expression} result: ${result}`);

        if (result) {
            call.callFlowIter.setRoot(this.getThenNode());
        } else {
            call.callFlowIter.setRoot(this.getElseNode());
        }
        return cb();
    }
}

module.exports = ConditionNode;

function getFn(call) {
    return {
        getChnVar: (varName) => {
           return call.getVar(varName);
        },
        getGlbVar: (varName) => {
            return call.getGlobalVar(varName);
        },
        match: (reg, val) => {
            let _reg = reg
                .replace(/\u0001/g, '\\')
                .match(new RegExp('^/(.*?)/([gimy]*)$'));

            if (!_reg) {
                _reg = [null, reg];
            }

            const _result = new RegExp(_reg[1], _reg[2]).exec(val);
            if (!_result)
                return false;

            call.regexpVariables.set(`${call.regexpVariables.size}`, _result);
            return true;
        },
        year: (param) => {
            return parseDate(param, call.getDate().years(), 9999);
        },
        yday: (param) => {
            return parseDate(param, call.getDate().dayOfYear(), 366);
        },
        mon: (param) => {
            return parseDate(param, (call.getDate().month() + 1), 12);
        },
        mday: (param) => {
            return parseDate(param, +call.getDate().format('D'), 31);
        },
        week: (param) => {
            return parseDate(param, +call.getDate().format('w'), 53);
        },
        mweek: (param) => {
            return parseDate(param, weekOfMonth(call.getDate()), 6);
        },
        wday: (param) => {
            return parseDate(param, +call.getDate().format('d') + 1, 7);
        },
        hour: (param) => {
            return parseDate(param, call.getDate().hours(), 23);
        },
        minute: (param) => {
            return parseDate(param, +call.getDate().format('m'), 59);
        },
        minute_of_day: (param) => {
            const now = call.getDate();
            return parseDate(param, (now.hours() * 60 + (+now.format('m')) ), 1440);
        },
        time_of_day: (param) => {
            param = param || '';
            const times = param.split(',');
            const now = call.getDate();
            const current = (now.hours() * 10000) + (+now.format('m') * 100) + (+now.format('s'));
            let _t;

            for (let i = 0, len = times.length; i < len; i++) {
                _t = times[i].split('-').map( (a) => pareTime(a));
                if ((current >= _t[0] && current <= _t[1])) 
                    return true;
            }
            return false;
        },
        limit: (params = '') => {
            const data = params.replace(/'/g, '').split(',');
            if (!data && !data[0]) {
                log.error('Bad parameters limit');
            }

            let result = call.getVar(`limit_usage_${call.domain}_${data[0]}`) || 0;
            if (data[1])
                return +result <= +data[1];
            else
                return +result;
        }
    }
}

//region private


/**
 *
 * @param expression
 * @param call
 * @returns {*}
 */
function execExpression(expression, call) {
    try {
        return new Function('sys, printError, module, process, global', 'try { return (' + expression + ') } catch (e) {printError(e)}')(call, printError)
    } catch (e) {
        log.error(e);
        return false;
    }
}

/**
 * 
 * @param e
 */
function printError(e) {
    log.error(e);
}

/**
 *
 * @param param
 * @param datetime
 * @param maxVal
 * @returns {boolean}
 */
function parseDate(param, datetime, maxVal) {
    param = param || '';
    const dateTimes = param.replace(/\s/g, '').split(',');
    let result = false;

    if (dateTimes[0] === "") {
        throw Error("bad parameters");
    }

    for (let i = 0; i < dateTimes.length; i++) {
        result = (dateTimes[i].indexOf('-') === -1)
            ? datetime == parseInt(dateTimes[i])
            : equalsDateTimeRange(datetime, dateTimes[i], maxVal);

        if (result === true) {
            return result
        }
    }
    return result;
}

/**
 * 
 * @param datetime
 * @param strRange
 * @param maxVal
 * @returns {boolean}
 */
function equalsDateTimeRange (datetime, strRange, maxVal) {
    let _min, _max;

    const dates = strRange.split('-');
    _min = parseInt(dates[0]);
    _max = dates[1]
        ? parseInt(dates[1])
        : maxVal;

    if (_min > _max)
        [_min, _max] = [_max, _min];

    return (datetime >= _min && datetime <= _max);
}

/**
 *
 * @param date
 * @returns {number}
 */
function weekOfMonth(m) {
    return m.week() - moment(m).startOf('month').week() + 1;
}

/**
 *
 * @param str
 * @returns {*}
 * @private
 */
function pareTime (str) {
    return str.split(':').reduce( (r, c, i) => {
        if (i === 0) {
            return +c * 10000
        } else if ( i === 1) {
            return r + (+c * 100)
        }
        return r + ( +c )
    } , 0);
}

//endregion