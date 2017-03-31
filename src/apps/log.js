/**
 * Created by igor on 27.03.17.
 */

"use strict";

module.exports = (acr) => {

    return function(call, cb) {
        call.execApp('log', `CONSOLE ${this.getArgs()}`, this.async);
        return cb()
    }
};