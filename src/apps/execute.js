/**
 * Created by igor on 31.03.17.
 */

"use strict";

module.exports = (acr) => {

    return function(call, cb) {
        const functionName = this.getArgs();
        const fn = call.callFlowIter.getFunction(functionName);
        if (!fn) {
            call.log(`Not found function ${functionName}`, true);
            return cb();
        }
        call.log(`Execute function ${functionName}`);

        const callIterator = call.callFlowIter;

        const end = (err, res) => {
            call.callFlowIter = callIterator;
            return cb(err, res);
        };

        call.callFlowIter = fn;

        const exec = (err, res) => {
            if (err)
                call.log(err, true);

            let app = fn.next() || fn.getParent();
            if (!app) {
                return end(err, res);
            }
            app.execute(call, (err, res) => {
                if (app.break === true) {
                    call.log(`Break call flow function ${functionName}`);
                    return end(err, res);
                }

                return exec(err, res);
            });

        };

        exec();
    }
};