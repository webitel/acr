/**
 * Created by igor on 29.03.17.
 */

"use strict";

const BaseNode = require('./baseNode'),
    log = require(__appRoot + '/lib/log')(module),
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

        const result = execExpression(this.expression, getFn());

        log.debug(`Expression ${this.expression} result: ${result}`);

        if (result) {
            call.callFlowIter.setRoot(this.getThenNode());
        } else {
            call.callFlowIter.setRoot(this.getElseNode());
        }
        return cb();
    }
}

module.exports = ConditionNode;



function getFn() {
    return {
        getChnVar: () => {},
        getGlbVar: () => {},
        match: () => {},
        year: () => {},
        yday: () => {},
        mon: () => {},
        mday: () => {},
        week: () => {},
        mweek: () => {},
        wday: () => {},
        hour: () => {},
        minute: () => {},
        minute_of_day: () => {},
        time_of_day: () => {},
        limit: () => {}
    }
}

function execExpression(expression, call) {
    try {
        return new Function('sys, module, process, global', 'try { return (' + expression + ') } catch (e) {}')(call)
    } catch (e) {
        log.error(e);
        return false;
    }
}