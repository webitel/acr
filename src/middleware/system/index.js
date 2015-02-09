/**
 * Created by i.navrotskyj on 30.01.2015.
 */

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module);
    conf = require('../../conf'),
    sysCollectionName = conf.get('mongodb:globalCollection'),
    globalVariables = {};

var sys = {
    getGlobalVariables: function (uuid, cb) {
        if (globalVariables[uuid]) {
            cb(null, globalVariables[uuid]);
            return;
        };

        this.getGlobalVarFromUUID(uuid, function (err, res) {
            if (err) {
                cb(err);
                return;
            };
            globalVariables[uuid] = res;
            cb(null, res);
        });
    },

    getGlobalVarFromUUID: function (uuid, cb) {
        try {
            if (!uuid || uuid == '') {
                cb(new Error('uuid is undefined'));
                return;
            }
            var gCollection = db.getCollection(sysCollectionName);
            gCollection.find({"Core-UUID": uuid})
                .limit(1)
                .toArray(cb);
        } catch (e) {
            cb(e);
        };
    }
};

module.exports = sys;