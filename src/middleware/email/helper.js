/**
 * Created by i.navrotskyj on 04.12.2015.
 */
'use strict';

var conf = require('../../conf'),
    db = require('../../lib/mongoDrv'),
    EMAIL_COLLECTION_NAME = conf.get("mongodb:emailCollection")
;

module.exports = {
    getSettings: function (domainName, cb) {
        try {
            var collection = db.getCollection(EMAIL_COLLECTION_NAME);

            return collection
                .findOne({"domain": domainName}, cb);
        } catch (e) {
            cb(e);
        }
    }
}