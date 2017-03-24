/**
 * Created by igor on 18.01.17.
 * TODO
 */

"use strict";

const api = require('./commands');
    
class CommandsMessage {
    constructor (msg = {properties: {}, fields: {}}) {
        this.properties = msg.properties;
        this.encodings = msg.properties.contentEncoding || 'utf8';

        this.api = null;
        this.args = null;

        this.exchange = this.properties.headers['x-api-resp-exchange'] || msg.fields.exchange;
        this.routingKey = this.properties.headers['x-api-resp-key'];

        const data = getJson(msg.content.toString(this.encodings));
        if (data) {
            this.api = data['exec-api'];
            this.args = data['exec-args'];
        }
    }

    execute (cb) {
        if (typeof this.api !== 'string')
            return cb(new Error('Bad api'));

        let fn = api;
        this.api.split('.').forEach( rout => {
            fn = fn && fn[rout];
        });

        if (typeof fn === 'function') {
            return cb(fn(this.args));
        } else {
            return cb(new Error(`No found api`));
        }
    }
}

module.exports = CommandsMessage;

function getJson(data = "") {
    try {
        return JSON.parse(data)
    } catch (e) {
        return {};
    }
}