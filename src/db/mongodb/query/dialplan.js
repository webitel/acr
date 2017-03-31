/**
 * Created by igor on 27.03.17.
 */

"use strict";

const conf = require(__appRoot + '/conf'),
    extensionCollection = conf.get('mongodb:extensionsCollection'),
    variablesCollection = conf.get('mongodb:variablesCollection'),
    publicCollection = conf.get('mongodb:publicCollection'),
    dialerCollection = conf.get('mongodb:dialerCollection'),
    defaultCollection = conf.get('mongodb:defaultCollection');
    
module.exports = db => {
    return {
        findExtension: (number, domain, cb) => {
            if (!number)
                return cb(new Error('destination_number is undefined'));

            db.collection(extensionCollection).findOne({
                "destination_number": number,
                "domain": domain,
                "disabled": {"$ne": true}
            }, cb);
        },

        findDefault: (domainName, cb) => {
            if (!domainName)
                return cb(new Error('Domain name is undefined'));

            db
                .collection(defaultCollection)
                .find({
                    domain: domainName,
                    disabled: {"$ne": true}
                })
                .sort({order: 1})
                .toArray(cb)
            ;

        }
    }
};