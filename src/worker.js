/**
 * Created by igor on 27.03.17.
 */

"use strict";

const path = require('path');
global.__appRoot = path.resolve(__dirname);
const acr = require('./acr')();
const log = require('./lib/log')(module);

process.title = "ACR V2";

process.on('SIGINT', function() {
    log.info('SIGINT received ...');
    acr.stop();
    process.exit(1);
    return true;
});

process.on('uncaughtException', function (err = {}) {
    log.error(err);
    acr.stop();
});

if (typeof gc === 'function') {
    setInterval(function () {
        gc();
        console.log('----------------- GC -----------------');
    }, 5000);
}