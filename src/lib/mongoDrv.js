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

mongoClient.connect(config.get('mongodb:uri') ,function(err, db) {
    if (err) {
        log.error('Connect db error: %s', err.message);
        throw err;
    };
    drv.dialplanCollection = db.collection(config.get("mongodb:collectionDialplan"));
    log.info('Connected db %s ', config.get('mongodb:uri'));
    db.on('close', function () {
        log.error('close mongo');
    })
});

module.exports = drv;