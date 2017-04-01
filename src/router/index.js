/**
 * Created by igor on 27.03.17.
 */

"use strict";

const CallTreeIterator = require('./iterator'),
    log = require(__appRoot + '/lib/log')(module),
    moment = require('moment-timezone'),
    MAP_VARIABLES = require(__appRoot + '/utils/mapVariables')
    ;

class Call {
    constructor (conn, schema, acr) {
        this._routeLog = [];
        this._id = conn._id;
        this._uuid = conn.channelData.getHeader('variable_uuid');
        this.localVariables = new Map();
        this.regexpVariables = new Map();

        const switchUUid = conn.channelData.getHeader('Core-UUID');

        if (!this._uuid) {
            this.log(`Not found uuid in ${this._id}`, true);
            this._uuid = this._id;
        }

        this.domain = schema.domain;
        this.timezone = schema.fs_timezone;
        this.callFlowIter = new CallTreeIterator(schema.callflow, acr);
        
        // this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');


        this.execApp = (appName, data, options = {}, cb) => {
            if (!appName)
                return cb(new Error('Application name is required.'));

            if (options.async) {
                this.log(`Execute async app: ${appName}, with data: ${data}`);
                conn.setEventLock(false);
            } else {
                this.log(`Execute sync app: ${appName}, with data: ${data}`);
                conn.setEventLock(true);
            }

            conn.execute(appName, data || '', cb);
        };

        this.getVar = varName => {
            return conn.channelData.getHeader('variable_' + varName)
                || conn.channelData.getHeader(varName)
                || this.getLocalVar(varName)
                || conn.channelData.getHeader(MAP_VARIABLES[varName])
                || '';
        };

        this.getLocalVar = varName => {
            if (this.localVariables.has(varName))
                return this.localVariables.get(varName);
            else
                return null;
        };

        this.getGlobalVar = varName => acr.getGlobalVar(switchUUid, varName);

        this.parseVarsFromText = (text = '') => {
            try {
                return text
                    .replace(/\$\$\{([\s\S]*?)\}/gi, (a, b) => {
                        return this.getGlobalVar(b);
                    })
                    .replace(/\$\{([\s\S]*?)\}/gi, (a, b) => {
                        return this.getVar(b)
                    });
            } catch (e) {
                log.error(`Parse variables error: `, text);
            };
        };

        this.updateChannelDump = dump => {
            if (dump) {
                dump.headers.forEach( (item) => {
                    conn.channelData.addHeader(item.name, item.value);
                });
            }
        };

        //region //Execute engine
        const end = () => {
            console.dir(this.logToJson(), {depth: 10, colors: true});
            // this.execApp('hangup', '');
            // return;
        };

        const exec = (err, res) => {
            if (err)
                this.log(err, true);

            let app = this.callFlowIter.next() || this.callFlowIter.getParent();
            if (!app) {
                return end();
            }
            app.execute(this, (err, res) => {
                if (app.break === true) {
                    this.log(`Break call flow`);
                    return end();
                }

                return exec(err, res);
            });

        };

        exec();

        //endregion engine

    }

    setLocalVar (varName, value) {
        if (!varName)
            return  false;

        this.localVariables.set(varName, value);
        return true;
    }

    getDate (momentDate) {
        if (!this.timezone) {
            return moment(momentDate).utc();
        }

        return moment(momentDate).tz(this.timezone);
    }

    log(data, e) {
        this._routeLog.push({
            time: Date.now(),
            log: data
        });

        if (e)
            log.error(data);
        else log.trace(`[${this._uuid}]: ${data}`); //TODO to uuid
    }

    logToJson () {
        return JSON.stringify(this._routeLog);
    }
}

module.exports = Call;

