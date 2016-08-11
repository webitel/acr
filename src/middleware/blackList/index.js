/**
 * Created by i.n. on 28.07.2015.
 */

let db = require('../../lib/mongoDrv'),
    conf = require('../../conf'),
    BL_COLLECTION_NAME = conf.get('mongodb:blackListCollection');

const BlackList = {
    check: function (domain, name, number, cb) {
        try {
            var collection = db.getCollection(BL_COLLECTION_NAME);
            collection.find({
                "domain": domain,
                "name": name,
                "number": number
            })
            .count(cb);
        } catch (e) {
            cb(e);
        }
    }
};

module.exports = BlackList;