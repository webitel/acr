/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var MongoDb = require("mongodb"),
    MongoClient = MongoDb.MongoClient,
    config = require('../conf'),
    log = require('../lib/log')(module)
    ;


class Drv {
    _initDB (db) {
        this.db = db;
        return this.db;
    }

    getCollection (name) {
        try {
            return this.db.collection(name)
        } catch (e) {
            log.error(`mongodb error: ${e.message}`);
        }
    }
}

let drv = new Drv(),
    timerId = null;

function connect () {
    if (timerId)
        clearTimeout(timerId);

    const mongodbClient = new MongoClient();

    const options = {
        autoReconnect: true,
        reconnectTries: Infinity,
        reconnectInterval: 1000
    };

    mongodbClient.connect(config.get('mongodb:uri'), options, function(err, db) {
        if (err) {
            log.error('Connect db error: %s', err.message);
            return timerId = setTimeout(connect, 1000);
        };
        drv._initDB(db);

        log.info('Connected db %s ', config.get('mongodb:uri'));
        db.on('close', function () {
            log.error('close mongo');
        });

        db.on('error', function (err) {
            log.error(err);
        });

        db.on('reconnect', function () {
            log.info('Reconnect MongoDB');
            drv._initDB(db);
        });
    });
}

connect();

module.exports = drv;