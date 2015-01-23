var winston = require('winston');
var conf = require('../conf');
require('winston-logstash');

function getLogger(module) {

    var path = process.pid + ':' + module.filename.split('//').slice(-2).join('//');

    var logLevels = {
        levels: {
            trace: 0,
            debug: 1,
            warn: 2,
            error: 3,
            info: 4
        },
        colors: {
            trace: 'yellow',
            debug: 'yellow',
            info: 'green',
            warn: 'yellow',
            error: 'red'
        }
    };
    winston.addColors(logLevels.colors);
    var logger = new (winston.Logger)({
        levels: logLevels.levels,
        transports: [
            new winston.transports.Console({
                colorize: true,
                level: conf.get('application:loglevel'),
                label: path,
                'timestamp': true
            })
        ]
    });
    if (conf.get('application:logstash:enabled')) {
        logger.add(winston.transports.Logstash, {
            port: conf.get('application:logstash:port'),
            node_name: conf.get('application:logstash:node_name'),
            host: conf.get('application:logstash:host'),
            level: conf.get('application:loglevel')
        });
    };
    return logger;
};

module.exports = getLogger;