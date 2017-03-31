/**
 * Created by igor on 27.03.17.
 */

"use strict";
    
const log = require(__appRoot + '/lib/log')(module),
    conf = require(__appRoot + '/conf'),
    MongoClient = require("mongodb").MongoClient,
    EventEmitter2 = require('eventemitter2').EventEmitter2;

class MongoDB extends EventEmitter2 {
    constructor () {
        super();
        this.client = new MongoClient();
        
        this._uri = conf.get('mongodb:uri');
        this._timerId = null;
        this.query = new Map();
        this.connect();
    }

    connect () {
        if (this._timerId)
            clearTimeout(this._timerId);


        const options = {
            autoReconnect: true,
            reconnectTries: Infinity,
            reconnectInterval: 1000
        };

        this.client.connect(this._uri, options, (err, db) => {
            if (err) {
                log.error(err);
                this._timerId = setTimeout(this.connect.bind(this), 1000);
                return;
            }

            log.info(`MongoDB connected to ${this._uri}`);

            db.on('error', err => {
                log.error(err);
                this.emit('error', err);
            });

            db.on('close', err => {
                log.error(`Close MongoDB connection...`);
                this.emit('error', err);
            });

            db.on('reconnect', () => { // TODO
                log.info(`Reconnect to ${this._uri}`);
                this.emit('connect', this);
            });

            this.query.set('dialplan', require('./query/dialplan')(db));

            this.emit('connect', this);
        });
    }

    getQuery (module, query) {
        if (!this.query.has(module)) {
            throw `No query module ${module}`;
        }

        if (query) {
            return this.query.get(module)[query]
        } else {
            return this.query.get(module);
        }
    }

    close () {
        //TODO
    }
}

module.exports = MongoDB;