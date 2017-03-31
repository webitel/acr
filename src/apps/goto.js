/**
 * Created by igor on 29.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module);

module.exports = (acr) => {

    return function(call, cb) {
        const tag = this.getArgs();
        call.log(`Goto ${tag}`);
        call.callFlowIter.goto(tag);
        return cb()
    }
};