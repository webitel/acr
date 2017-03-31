/**
 * Created by igor on 29.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module);

module.exports = (acr) => {

    return function(call, cb) {
        const tag = this.getArgs();

        //TODO ADD old goto support - delete new version;
        // isNaN(tag) &;

        if (call.callFlowIter.goto(tag)) {
            call.log(`Goto ${tag}`);
        } else {
            call.log(`Goto not found ${tag}`, true);
        }

        return cb()
    }
};