/**
 * Created by Igor Navrotskyj on 23.09.2015.
 */

'use strict';
var log = require('../../lib/log')(module),
    conf = require('../../conf'),
    db = require('../../lib/mongoDrv'),
    LOCATION_COLLECTION_NAME = conf.get("mongodb:locationNumberCollection")
    ;

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
            number = number.replace(/\D/g, '');
            
            var numbers = [];
            number.split('').reduce(function(r, v, i, a) {
                numbers.push(a.slice(0, i).join(''));
                return r + v
            });

            collection
                .find({"sysLength": number.length,"code": { $in: numbers}})
                .sort({"sysOrder": -1})
                .limit(1)
                .toArray(function (err, array) {
                    if (err) {
                        if (cb) cb();
                        return log.error(err);
                    };

                    var _array = array[0];
                    var goecode =  _array && _array.goecode && _array.goecode[0];
                    if (_array && goecode && goecode['latitude'] && goecode['longitude']) {
                        var locStr = ''.concat(goecode['latitude'], ', ', goecode['longitude']);
                        var _vars = [
                            'webitel_location=' + locStr,
                            'webitel_location_country=' + _array['country'],
                            'webitel_location_type=' + _array['type']
                        ];
                        if (_array['city']) {
                            _vars.push("webitel_location_city=" + _array['city']);
                        };

                        if (goecode['countryCode']) {
                            _vars.push('webitel_location_country_code=' + goecode['countryCode']);
                        };

                        log.debug('Number %s location: %s', number, goecode.formattedAddress);
                        scope.__setVar({
                            "setVar": _vars
                        }, cb);
                    }
                });

        } catch (e) {
            log.error(e);
            if (cb)
                cb();
        };
    };
};