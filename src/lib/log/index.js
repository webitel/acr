/**
 * Created by igor on 27.03.17.
 */

"use strict";

var winston = require('winston');
var conf = require(__appRoot + '/conf');

function getLogger(module) {
    let pathDirectory = module.filename.split(/\/+/).slice(-3);
    let path = pathDirectory.join('\\') + '(' + process.pid + ')';

    let logLevels = {
        levels: {
            trace: 4,
            debug: 3,
            warn: 2,
            error: 1,
            info: 0
        },
        colors: {
            trace: 'cyan',
            debug: 'white',
            info: 'green',
            warn: 'yellow',
            error: 'red'
        }
    };

    return new (winston.Logger)({
        levels: logLevels.levels,
        colors: logLevels.colors,
        transports: [
            new winston.transports.Console({
                colorize: 'all',
                level: conf.get('application:loglevel'),
                label: path,
                timestamp: false
            })
        ]
    });
}

module.exports = getLogger;