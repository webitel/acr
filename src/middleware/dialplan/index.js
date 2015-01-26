/**
 * Created by i.navrotskyj on 26.01.2015.
 */

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module);

var dialplan = {
    findActualDialplan: function (number, cb) {
        if (!number || number == '') {
            cb(new Error('destination_number is undefined'));
            return;
        }
        var dialCollection = db.dialplanCollection;
        dialCollection.find({"destination_number": number})
            .sort({"version": -1})
            .limit(1)
            .toArray(cb);
    }
};

module.exports = dialplan;