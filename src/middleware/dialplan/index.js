/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    publicCollection = config.get('mongodb:publicCollection'),
    defaultCollection = config.get('mongodb:defaultCollection'),
    extensionCollection = config.get('mongodb:extensionsCollection'),
    variablesCollection = config.get('mongodb:variablesCollection')
    ;

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

    findDomainVariables: function (domainName, cb) {
        var collection = db.getCollection(variablesCollection);
        collection.findOne({
            "domain": domainName
        }, function (err, res) {
            if (err) {
                log.error(err['message']);
            };
            if (cb)
                cb(err, res);
        });
    },

    updateDomainVariables: function (domainName, variables, cb) {
        try {
            var doc = {
                    "variables": variables,
                    "domain": domainName
                },
                collection = db.getCollection(variablesCollection)
                ;

            collection.update({
                    "domain": domainName
                },
                doc,
                {upsert: true},
                cb
            );
        } catch (e) {
            cb(e);
        }
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
    },

    findActualExtension: function (number, domain, cb) {
        if (!number || number == '') {
            cb(new Error('destination_number is undefined'));
            return;
        };
        var dialCollection = db.getCollection(extensionCollection);
        dialCollection.findOne({
                "destination_number": number,
                "domain": domain
            }, cb);
    }
};

module.exports = dialplan;