/**
 * Created by igor on 27.03.17.
 */

"use strict";
    
const EventEmitter2 = require('eventemitter2').EventEmitter2,
    DB = require('./db'),
    ESL = require('./lib/modesl'),
    conf = require('./conf'),
    context = require('./context'),
    parseApi = require('./utils/parser').api,
    fs = require('fs'),
    log = require('./lib/log')(module);

const DEFAULT_HANGUP_CAUSE = 'DESTINATION_OUT_OF_ORDER';

class ACR extends EventEmitter2 {
    constructor () {
        super();
        this.db = new DB();
        this._server = null;

        this.db.once('connect', () => {
            this.createServer();
        });

        this.globalVar = new Map();

        this.apps = new Map();
        this._appNames = ['if', 'case', 'switch', 'break'];

        this.registerApplications();
    }

    getApplication (name) {
        return this.apps.get(name)
    }

    registerApplications () {
        const appsFiles = fs.readdirSync(__appRoot + '/apps');

        appsFiles.forEach( fileName => {
            const appName = fileName.replace(/(.*)\.(.*)?/, '$1');
            log.info(`Register application ${appName}`);
            const app = require(`./apps/${fileName}`)(this);

            if (typeof app !== 'function')
                throw `Bad application file ${fileName}`;

            this.apps.set(appName, app);
            this._appNames.push(appName);
        });
    }

    existsApp (name) {
        return ~this._appNames.indexOf(name);
    }
    
    initGlobalVar (switchUuid, conn, cb) {
        if (this.globalVar.has(switchUuid)) {
            return cb(null);
        }

        conn.api('global_getvar', res => {
            const vars = parseApi(res && res.getBody());
            this.globalVar.set(switchUuid, vars);
            log.info(`Set global variables from ${switchUuid}`);
            return cb(null);
        });
    }

    getGlobalVar (switchUuid, name) {
        if (this.globalVar.has(switchUuid)) {
            return this.globalVar.get(switchUuid)[name]
        }
        log.warn(`No global var ${switchUuid}`);
        return null;
    }

    createServer () {
        const host = conf.get('server:host');
        const port = process.env['WORKER_PORT'] || 10030;
        this._server = new ESL.Server(
            {
                host,
                port,
                myevents: true
            },
            () => log.info(`Open esl server on ${host}:${port}`)
        );
        this._server.on('error', this.onError.bind(this));
        this._server.on('connection::open', this.onOpenConnection.bind(this));
        this._server.on('connection::close', this.onCloseConnection.bind(this));
        this._server.on('connection::ready', this.onReadyConnection.bind(this));

    }

    closeConnection (conn, err) {
        if (err)
            log.error(err);
        conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
    }

    onReadyConnection (conn, id) {
        try {
            context(this, conn, id);
        } catch (e) {
            log.error(`Route ${id} error: `, e);
        }
    }

    onOpenConnection (conn, id, allCount) {
        log.debug(`Open connection ${id} [all connection count ${allCount}]`);
        conn.on('error', (err) => {
            log.warn(`Call ${id} error: `, err);
        });
    };

    onCloseConnection (conn, id, allCount) {
        log.debug(`Close connection ${id} [all connection count ${allCount}]`);
    }

    onError (err) {
        log.error(err);
    }


    stop (e) {
        log.info('stop');
        this.db.close();
        process.exit(1);
    }
}

const acr = new ACR();

function getFnName(cond) {
    if (!cond)
        return null;

    var propKeys = Object.keys(cond);
    if (propKeys.length === 1) {
        return propKeys[0];
    } else if (propKeys.length === 0) {
        return null;
    } else {
        for (var i = 0, len = propKeys.length; i < len; i++) {
            if (propKeys[i] !== 'break' && propKeys[i] !== 'async' && propKeys[i] !== 'tag') {
                return propKeys[i];
            }
        }
    }
}

module.exports = () => acr;