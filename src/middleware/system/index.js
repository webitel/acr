/**
 * Created by i.navrotskyj on 30.01.2015.
 */
var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module);
    conf = require('../../conf'),
    sysCollectionName = conf.get('mongodb:globalCollection');

var sys = {
    getGlobalVarFromUUID: function (uuid, cb) {
        try {
            if (!uuid || uuid == '') {
                cb(new Error('uuid is undefined'));
                return;
            }
            var gCollection = db.getCollection(sysCollectionName);
            gCollection.find({"Core-UUID": uuid})
                .sort({"version": -1})
                .limit(1)
                .toArray(cb);
        } catch (e) {
            cb(e);
        }
    }
};

module.exports = sys;