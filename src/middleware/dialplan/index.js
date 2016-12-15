/**
 * Created by i.navrotskyj on 26.01.2015.
 */

let db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    ObjectID = require('mongodb').ObjectID,
    publicCollection = config.get('mongodb:publicCollection'),
    dialerCollection = config.get('mongodb:dialerCollection'),
    defaultCollection = config.get('mongodb:defaultCollection'),
    extensionCollection = config.get('mongodb:extensionsCollection'),
    variablesCollection = config.get('mongodb:variablesCollection')
    ;

const dialplan = {
    findActualPublicDialplan: function (number, cb) {
        if (!number || number == '') {
            cb(new Error('destination_number is undefined'));
            return;
        }
        let dialCollection = db.getCollection(publicCollection);
        dialCollection.find({"destination_number": number, "disabled": {"$ne": true}})
            .sort({"version": -1})
            .limit(1)
            .toArray(cb);
    },

    findDomainVariables: function (domainName, cb) {
        let collection = db.getCollection(variablesCollection);
        collection.findOne({
            "domain": domainName
        }, function (err, res) {
            if (err) {
                log.error(err['message']);
            }
            if (cb)
                cb(err, res);
        });
    },

    updateDomainVariables: function (domainName, variables, cb) {
        try {
            let doc = {
                    "variables": variables,
                    "domain": domainName
                },
                collection = db.getCollection(variablesCollection);

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

    findDialerDialplan: function (id, domainName, cb) {
        if (ObjectID.isValid(id))
            id = new ObjectID(id);

        let dialCollection = db.getCollection(dialerCollection);

        dialCollection.findOne({"_id": id}, {_cf: 1, amd: 1}, cb);
    },
    
    findActualDefaultDialplan: function (domainName, cb) {
        if (!domainName || domainName == '') {
            cb(new Error('domain is undefined'));
            return;
        }
        let dialCollection = db.getCollection(defaultCollection);
        dialCollection.find({"domain": domainName, "disabled": {"$ne": true}}, {read  : true})
            .sort({"order": 1})
            .toArray(cb);
    },

    findActualExtension: function (number, domain, cb) {
        if (!number || number == '') {
            cb(new Error('destination_number is undefined'));
            return;
        }
        let dialCollection = db.getCollection(extensionCollection);
        dialCollection.findOne({
                "destination_number": number,
                "domain": domain,
                "disabled": {"$ne": true}
            }, cb);
    }
};

module.exports = dialplan;