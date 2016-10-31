/**
 * Created on 02.04.2015.
 */
'use strict';

let db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    moment = require('moment-timezone'),
    CALENDAR_COLLECTION = config.get('mongodb:calendarCollection');

module.exports = function (application, cb) {
    try {
        let prop = application && application.calendar,
            scope = this;

        if (!prop) {
            log.error("Bad calendar application");
            return cb && cb();
        }

        let name = prop.name,
            varName = prop.setVar;

        if (!name || !varName) {
            log.error('Bad calendar property');
            return cb && cb();
        }

        var callback = function (err, ok) {
            if(err)
                log.error(err);

            scope.__setVar({
                "setVar": `${varName}=${ok || false}`
            }, cb);
        };

        let calendarCollection = db.getCollection(CALENDAR_COLLECTION);
        calendarCollection.findOne(
            {name: name, domain: this.domain},
            (err, calendar) => {
                if (err)
                    return callback(err);

                if (!calendar) {
                    return callback(new Error(`Not found calendar ${name}`), false);
                }

                let current;
                if (this.offset)
                    current = moment().tz(this.offset);
                else if (calendar.timeZone && calendar.timeZone.id)
                    current = moment().tz(calendar.timeZone.id);
                else current = moment();

                let currentTime = current.valueOf();

                // Check range date;
                if (calendar.startDate && currentTime < calendar.startDate)
                    return callback(null, false);
                else if (calendar.endDate && currentTime > calendar.endDate)
                    return callback(null, false);

                //Check work
                let isAccept = false;

                if (calendar.accept instanceof Array) {
                    let currentTimeOfDay = current.get('hours') * 60 + current.get('minutes'),
                        currentWeek = current.isoWeekday()
                        ;

                    for (let i = 0, len = calendar.accept.length; i < len; i++) {
                        isAccept = currentWeek === calendar.accept[i].weekDay && between(currentTimeOfDay, calendar.accept[i].startTime, calendar.accept[i].endTime);
                        if (isAccept)
                            break;
                    }

                } else {
                    // TODO ERROR ???
                    return callback(new Error('Bad record ?'));
                }

                if (!isAccept)
                    return callback(null, false);

                // Check holiday
                if (calendar.except instanceof Array) {
                    let currentDay = current.get('date'),
                        currentMonth = current.get('month'),
                        currentYear = current.get('year'),
                        exceptDate
                        ;

                    for (let i = 0, len = calendar.except.length; i < len; i++) {
                        exceptDate = moment(calendar.except[i].date);
                        if (exceptDate.get('date') == currentDay && exceptDate.get('month') == currentMonth &&
                                (calendar.except[i].repeat === 1 || (calendar.except[i].repeat === 0 && exceptDate.get('year') == currentYear)) )
                            return callback(null, false);
                    }
                }

                return callback(null, true);
            }
        );

    } catch (e) {
        log.error(e['message']);
        if (cb)
            cb(e);
    }
};

function between(x, min, max) {
    return x >= min && x <= max;
}