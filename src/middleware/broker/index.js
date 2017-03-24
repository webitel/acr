/**
 * Created by igor on 02.08.16.
 */

"use strict";

const Amqp = require('amqplib'),
    log = require('../../lib/log')(module),
    conf = require('../../conf'),
    EventEmitter2 = require('eventemitter2').EventEmitter2;

    
class RPC extends EventEmitter2 {
    constructor (config= {}) {
        super();
        if (!config.hasOwnProperty('connectionString'))
            return log.error(`No broker configure connectionString`);

        this._conf = config;
        this._timerId = null;
        this.connection = null;
        this.channel = null;
        this.queue = null;
        this._eventsExchanges = [];
        this._comandsExchanges = [];
        this._comandsQueue = null;

        this.connect();
    }

    connect () {
        log.debug(`Try connect to ${this._conf.connectionString}`);
        if (this._timerId)
            clearTimeout(this._timerId);

        Amqp
            .connect(this._conf.connectionString)
            .then(conn => {

                conn.on('error', (err) => {
                    if (err.message !== "Connection closing") {
                        return log.error("conn error", err);
                    }
                    conn.close();
                });

                conn.on('close', e => {
                    this.onConnectError(e)
                });

                conn
                    .createConfirmChannel()
                    .then(channel => {

                        this.onConnect(conn, channel)
                    })
                    .catch(this.onConnectError.bind(this));
            })
            .catch(this.onConnectError.bind(this))
    }

    onConnectError (err = {}) {
        console.dir(err.stack);
        this.connection = null;
        this.channel = null;
        this._eventsExchanges = [];
        this._comandsExchanges = [];
        this._comandsQueue = null;
        log.error(err.message);
        this.emit('disconnect', this);
        log.warn(`Try reconnect amqp`);
        this._timerId = setTimeout(this.connect.bind(this), 1000);
    }

    onConnect (connection, channel) {
        this.connection = connection;
        this.channel = channel;
        log.info(`Success connect to ${this._conf.connectionString}`);
        this.initExchangeEvent();
        this.initExchangeCommands();
        this.emit('connect', this);
    }

    initExchangeEvent (i = 0) {
        if (this._conf.events && this._conf.events[i]) {
            const ex = this._conf.events[i];
            if (!ex.name) {
                log.warn(`Bad config event exchange: ${ex}`);
                return this.initExchangeEvent(++i);
            }
            this
                .channel
                .assertExchange(ex.name, ex.type || "topic", {durable: true})
                .then(e => {
                    log.info(`Init exchange ${ex.name} - success`);
                    this._eventsExchanges.push({
                        name: ex.name,
                        format: ex.format + '' ? ex.format.split(',') : []
                    });
                    return this.initExchangeEvent(++i);
                })
                .catch(e => {
                    log.error(`Bad exchange ${ex.name}:`, e);
                    return this.initExchangeEvent(++i);
                })
        } else {
            log.debug('End init events exchanges');
        }
    }

    initExchangeCommands () {
        if (this._conf.commands instanceof Array) {
            this
                .channel
                .assertQueue('', {autoDelete: true, durable: false, exclusive: true})
                .then(qok => {

                    this
                        .channel
                        .consume(qok.queue, msg => this._onCommand(msg), {noAck: true});
                    this._comandsQueue = qok.queue;
                    this.bindExchangeCommands();
                })
                .catch(e => {
                    log.error(`Failed create commands queue`);
                    this.initExchangeCommands();
                })
        } else {
            log.debug('End init commands exchanges');
        }
    }

    bindExchangeCommands (i = 0) {
        if (this._conf.commands && this._conf.commands[i]) {
            const ex = this._conf.commands[i];
            if (!ex.name) {
                log.warn(`Bad config commands exchange: ${ex}`);
                return this.bindExchangeCommands(++i);
            }
            this
                .channel
                .assertExchange(ex.name, ex.type || "topic", {durable: true})
                .then(e => {
                    log.info(`Init exchange ${ex.name} - success`);
                    this._comandsExchanges.push({
                        name: ex.name,
                        format: ex.format + '' ? ex.format.split(',') : []
                    });
                    this.channel.bindQueue(this._comandsQueue, ex.name, ex.format);
                    return this.bindExchangeCommands(++i);
                })
                .catch(e => {
                    log.error(`Bad exchange ${ex.name}:`, e);
                    return this.bindExchangeCommands(++i);
                })
        } else {
            log.debug('End init commands exchanges');
        }
    }

    _onCommand (msg) {
        if (!msg)
            return;

        const apiMsg = new ApiMsg(msg);
        log.debug(`exec ${apiMsg.api}, args: `, apiMsg.args);
        apiMsg.execute( (res) => {
            return this.sendCommandsResponse(apiMsg, JSON.stringify(res));
        });
    }

    sendCommandsResponse (apiMsg, data) {
        if (apiMsg.properties.replyTo && apiMsg.properties.correlationId) {
            this
                .channel
                .sendToQueue(apiMsg.properties.replyTo, new Buffer(data), {correlationId: apiMsg.properties.correlationId});
        } else if (apiMsg.routingKey) {
            this
                .channel
                .publish(apiMsg.exchange, apiMsg.routingKey, new Buffer(data));
        }

    }

    checkExchange (ex) {
        return this.channel.checkExchange(ex)
    }

    sendEvent (event) {
        log.trace(`Send event ${event.type}`);
        this._eventsExchanges.forEach( e => {
            this
                .channel
                .publish(e.name, event.parseRk(e.format), new Buffer(event.serialize('json')))
        })
    }
}
const ApiMsg = require('./commandsMsg');
const rpc = new RPC(conf.get('broker'));

module.exports = rpc;