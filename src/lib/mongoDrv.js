/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var MongoDb = require("mongodb")
    , MongoClient = MongoDb.MongoClient
    , format = require('util').format
    , config = require('../conf')
    , log = require('../lib/log')(module);

var mongoClient = new MongoClient();


var drv = function (option) {

};

drv._initDB = function (db) {
    this.db = db;
    return this.db;
};


mongoClient.connect(config.get('mongodb:uri') ,function(err, db) {
    if (err) {
        log.error('Connect db error: %s', err.message);
        throw err;
    };
    drv._initDB(db);

    log.info('Connected db %s ', config.get('mongodb:uri'));
    db.on('close', function () {
        log.error('close mongo');
    });

    db.on('error', function (err) {
        log.error(err);
    });
});

drv.getCollection = function (name) {
    try {
        return this.db.collection(name)
    } catch (e) {
        log.error(e.message);
    }
};

module.exports = drv;