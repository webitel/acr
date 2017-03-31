/**
 * Created by igor on 27.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module)
    ;
    
module.exports = (acr) => {
    
    function a(args = {}) {
        return {
            expression: args.sysExpresion
            
        }
    }
    
    return (call, data, options = {}, cb) => {
        const {sysExpression, then: thenApps, else: elseApp} = data;

        if (!sysExpression)
            return cb(new Error(`Bad application expression`));

        log.trace(`Parse expression: ${sysExpression}`);
        const result = execExpression(sysExpression, getFn());

        log.trace(`Expression ${sysExpression} result: ${result}`);

        if (result) {
            if (thenApps instanceof Array) {
                call.callFlowIter.setRoot(thenApps);
                return cb()
            }

            return cb(new Error(`No [then] application in if`));
        } else {
            if (elseApp instanceof Array) {
                call.callFlowIter.setRoot(elseApp);
                return cb()
            }
            return cb(new Error(`No [else] application in if`));
        }
    }
};

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