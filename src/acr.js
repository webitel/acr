/**
 * Created by igor on 23.03.17.
 * //TODO
 */

"use strict";


const esl = require('./lib/modesl'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    publicContext = require('./middleware/publicContext'),
    defaultContext = require('./middleware/defaultContext'),
    dialerContext = require('./middleware/dialerContext'),
    globalCollection = require('./middleware/system'),
    DEFAULT_HANGUP_CAUSE = require('./const').DEFAULT_HANGUP_CAUSE,
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    PUBLIC_CONTEXT = 'public'
    ;
    
class Acr extends EventEmitter2 {
    constructor() {
        super();

        const eslServer = this.server = new esl.Server(
            {
                host: conf.get('server:host'),
                port: process.env['WORKER_PORT'] || 10030,
                myevents: true
            },
            function() {
                log.info("ESL server is up port " + this.port);
            }
        );

        eslServer.on('connection::open', (conn, id) => {
            conn.on('error', function (error) {
                log.warn('Call %s error: %s', id, error.message);
            });
        });

        eslServer.on('connection::ready', function(conn, id, allCountSocket) {

            log.trace('New call %s [all socket: %s]', id, allCountSocket);

            let lastExecuteDump;
            conn.on('esl::end', () => {
                if (conn && conn.__callRouter) {
                    var end = () => {
                        if (conn.__callRouter) {
                            conn.__callRouter.stop();
                            delete conn.__callRouter;
                        }
                    };

                    if (conn.__callRouter.onDisconnectCallflow instanceof Array && conn.__callRouter.onDisconnectCallflow.length > 0) {
                        try {
                            conn.__callRouter._updateChannelDump(lastExecuteDump);
                            conn.__callRouter.end = false;
                            conn.__callRouter.execute(conn.__callRouter.onDisconnectCallflow, () => {
                                log.trace(`end onDisconnectCallflow`);
                                end();
                            })
                        } catch (e) {
                            log.error(e);
                            end();
                        }
                    } else {
                        end();
                    }
                }
            });

            conn.on(`esl::event::CHANNEL_EXECUTE_COMPLETE::*`, (e) => {
                lastExecuteDump = e;
            });

            try {
                var context = conn.channelData.getHeader('Channel-Context'),
                    dialerId = conn.channelData.getHeader('variable_dlr_queue'),
                    destinationNumber = conn.channelData.getHeader('Channel-Destination-Number') ||
                        conn.channelData.getHeader('Caller-Destination-Number') || conn.channelData.getHeader('variable_destination_number');

                log.debug('Call %s -> %s ->', id, destinationNumber, context);

                globalCollection.getGlobalVariables(conn, conn.channelData.getHeader('Core-UUID'), function (err, globalVariable) {
                    if (err) {
                        log.error(err.message);
                        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
                        return
                    }

                    var soundPref = '\/$${sounds_dir}\/en\/us\/callie';
                    if (conn.channelData.getHeader('variable_default_language') == 'ru') {
                        soundPref = '\/$${sounds_dir}\/ru\/RU\/elena';
                    }

                    conn.execute('set', 'sound_prefix=' + soundPref);

                    if (context === PUBLIC_CONTEXT) {
                        publicContext(conn, destinationNumber, globalVariable, !conn.channelData.getHeader('variable_webitel_direction'));
                    } else if (dialerId) {
                        dialerContext(conn, dialerId, globalVariable, !conn.channelData.getHeader('variable_webitel_direction'));
                    } else {
                        defaultContext(conn, destinationNumber, globalVariable, !conn.channelData.getHeader('variable_webitel_direction'));
                    }

                });

            } catch (e) {
                log.error(e.message);
                conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            }

        });

        eslServer.on('error', function (err) {
            log.error(err);
        });

        eslServer.on('connection::close', function(c, id, allCount) {
            log.trace("Call end %s [all socket: %s]", id, allCount);
        });
    }

    getConnection (id) {
        if (this.server.connections.hasOwnProperty(id)) {
            return this.server.connections[id];
        }
        return null;
    }
}

global.application = new Acr();