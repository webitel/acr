/**
 * Created by igor on 28.06.16.
 */

"use strict";

const log = require('../../lib/log')(module),
    Event = require('modesl').Event
    ;

module.exports = (CallRouter, appName) => {
    CallRouter.prototype['__' + appName] = function (app, cb) {
        let prop = app[appName],
            to = this._parseVariable(prop.to),
            message = this._parseVariable(prop.message)
            ;

        if (!to || !message) {
            log.warn(`Bad telegram parameters`);
            return cb && cb();
        }

        let event = new Event('CUSTOM', 'webitel::telegram');

        event.addHeader('Channel-Presence-ID', `${to.split('@')[0]}@${this.domain}`);
        event.addHeader('domain', this.domain);
        event.addHeader('type', 'text/plain');
        event.addHeader('Content-Type', 'text/plain');
        event.addBody(message);

        this.connection.sendEvent(event, (e, r) => {
            log.trace(`Send telegram notification: ${r && r["Reply-Text"]}`);
        });
        return cb && cb();
    } 
};