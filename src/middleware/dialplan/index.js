/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    publicCollection = config.get('mongodb:publicCollection'),
    defaultCollection = config.get('mongodb:defaultCollection');

var dialplan = {
    findActualPublicDialplan: function (number, cb) {
        if (!number || number == '') {
            cb(new Error('destination_number is undefined'));
            return;
        }
        var dialCollection = db.getCollection(publicCollection);
        dialCollection.find({"destination_number": number})
            .sort({"version": -1})
            .limit(1)
            .toArray(cb);
    },
    
    findActualDefaultDialplan: function (domainName, cb) {
        if (!domainName || domainName == '') {
            cb(new Error('domain is undefined'));
            return;
        }
        var dialCollection = db.getCollection(defaultCollection);
        dialCollection.find({"domain": domainName}, {read  : true})
            .sort({"order": 1})
            .toArray(cb);
    }
};

module.exports = dialplan;