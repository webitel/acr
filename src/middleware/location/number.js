/**
 * Created by Igor Navrotskyj on 23.09.2015.
 */

'use strict';
var log = require('../../lib/log')(module),
    conf = require('../../conf'),
    db = require('../../lib/mongoDrv'),
    LOCATION_COLLECTION_NAME = conf.get("mongodb:locationNumberCollection")
    ;

'use strict';

module.exports = function (CallRouter, applicationName) {

    CallRouter.prototype.__geoLocation = function (app, cb) {
        var prop = app[applicationName],
            varName = prop['variable'],
            pattern = prop['regex'],
            result = prop['result'],
            scope = this,
            number = this.channelDestinationNumber;
        ;

        if (varName) {
            number = this.getChnVar(varName);
        };

        if (!number) {
            log.error('Number not found.');
            if (cb)
                return cb();
        };

        try {
            if (pattern && result) {
                var _r = pattern.match(new RegExp('^/(.*?)/([gimy]*)$'));
                // Bad destination reg exp value
                if (!_r) {
                    _r = [null, pattern]
                };
                try {
                    number = number.replace(new RegExp(_r[1], _r[2]), result);
                } catch (e) {
                    log.warn(e.message);
                };
            };

            var collection = db.getCollection(LOCATION_COLLECTION_NAME);
            collection
                .find({ "$where": "obj.sysSearch.test('" + number +"')" })
                .sort({"type": 1, "code": -1})
                .limit(1)
                .toArray(function (err, array) {
                    if (err) {
                        return log.error(err);
                    };

                    var goecode = array[0] && array[0].goecode && array[0].goecode[0];
                    if (goecode && goecode['latitude'] && goecode['longitude']) {
                        var locStr = ''.concat(goecode['latitude'], ', ', goecode['longitude']);
                        log.debug('Number %s location: %s', number, goecode.formattedAddress);
                        scope.__setVar({
                            "setVar": 'webitel_location=' + locStr
                        }, cb);
                    } else {
                        if (cb)
                            return cb();
                    }
                });

        } catch (e) {
            log.error(e);
            if (cb)
                cb();
        };

    };
};