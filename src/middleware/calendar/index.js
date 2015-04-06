/**
 * Created on 02.04.2015.
 */

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    CALENDAR_COLLECTION = config.get('mongodb:calendarCollection');

var METHOD = {
    IN: "in",
    NIN: "nin"
};

module.exports = function (conn, app, cb) {
    try {
        var calendarCollection = db.getCollection(CALENDAR_COLLECTION),
            time = conn.DateOffset().getTime(),
            tags = app['tags'],
            limit = app['limit'] || 1
            ;

        var _q = {
            "$and": [
                {
                    "domain": conn['domain']
                },
                {
                    "$and": [{
                        "startDate": {
                            "$lte": time
                        }
                    },
                    {
                        "dueDate": {
                            "$gte": time
                        }
                    }]
                }
            ]
    };

    if (tags) {
        var _tagsQuery = {
                "tags": {}
            },
            _method = app['method'] === METHOD.NIN
                ? "$nin"
                : "$in"
            ;
        if (tags instanceof Array) {
            _tagsQuery["tags"][_method] = tags
        } else if (typeof tags === 'string') {
            _tagsQuery["tags"][_method] = [tags]
        };
        _q['$and'].push(_tagsQuery);

    };
    calendarCollection.find(_q, {read  : true})
            .limit(limit)
            .toArray(cb);

    } catch (e) {
        log.error(e['message']);
        if (cb)
            cb(e);
    }
};