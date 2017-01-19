/**
 * Created by igor on 18.01.17.
 */

"use strict";
    
const EslEvent = require('./modesl').Event,
    conf = require('../conf'),
    HOST_NAME = conf.get('server:host');

class Event extends EslEvent {
    constructor (type, subclass) {
        super(type, subclass);
        this.addHeader('ACR-Hostname', HOST_NAME);
    }

    parseRk (format = []) {
        return format.map( i => {
            let h = this.getHeader(i);
            return encodeRK(h || "");
        }).join('.')
    }
}

module.exports = Event;


function encodeRK (rk) {
    try {
        if (rk)
            return encodeURIComponent(rk)
                .replace(/\.|\:/g, v => {
                    if (v === '.') {
                        return '%2E'
                    } else if (v === ':') {
                        return '%3A'
                    }
                    return v;
                });
    } catch(e) {
        return '';
    }
}