/**
 * Created by igor on 27.03.17.
 */

"use strict";
    
module.exports = (acr) => {

    return function(call, cb) {
        let app = null;
        const data = this.getArgs();

        if (!data || /\b200\b|\bOK\b/i.test(data)) {
            app = "answer";
        } else if (/\b183\b|\bSession Progress\b/i.test(data)) {
            app = "pre_answer";
        } else if (/\b180\b|\bRinging\b/i.test(data)) {
            app = "ring_ready"
        }

        if (!app) {
            return cb(new Error(`Bad answer value: ${data}`));
        }

        call.execApp(app, "", {async: this.async}, dump => {
            return cb(null, dump);
        });
    }
};