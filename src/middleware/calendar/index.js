/**
 * Created on 02.04.2015.
 */
'use strict';

var db = require('../../lib/mongoDrv'),
    log = require('../../lib/log')(module),
    config = require('../../conf'),
    moment = require('moment-timezone'),
    CALENDAR_COLLECTION = config.get('mongodb:calendarCollection');

var METHOD = {
    IN: "in",
    NIN: "nin"
};

module.exports = function (application, cb) {
    try {
        let prop = application && application.calendar;

        if (!prop) {
            log.error("Bad calendar application");
            return cb && cb();
        };

        let name = prop.name,
            varName = prop.setVar;

        if (!name || !varName) {
            log.error('Bad calendar property');
            return cb && cb();
        };

        let calendarCollection = db.getCollection(CALENDAR_COLLECTION);
        calendarCollection.findOne(
            {name: name, domain: this.domain},
            (err, calendar) => {
                if (err) {
                    log.error(err);
                    return cb && cb(err);
                };

                if (!calendar) {
                    log.trace(`Not found calendar ${name}`);
                    return cb && cb();
                };

                let current;
                if (this.offset)
                    current = moment().tz(this.offset);
                else if (res.timeZone && res.timeZone.id)
                    current = moment().tz(res.timeZone.id);
                else current = moment();
                
                var callback = function (err, ok) {
                    if(err)
                        log.error(err);

                    return cb && cb(err);
                };

                let currentTime = current.valueOf();

                // Check range date;
                if (calendar.startDate && currentTime < calendar.startDate)
                    return callback();
                else if (calendar.endDate && currentTime > calendar.endDate)
                    return callback();

                //Check work
                let isAccept = false;

                if (calendar.accept instanceof Array) {
                    let currentTimeOfDay = current.get('hours') * 60 + current.get('minutes'),
                        currentWeek = current.day()
                        ;

                    for (let i = 0, len = calendar.accept.length; i < len; i++) {
                        isAccept = between(currentTimeOfDay, calendar.accept[i].startTime, calendar.accept[i].endTime);
                        if (isAccept)
                            break;
                    };

                } else {
                    // TODO ERROR ???
                    return callback(new Error('Bad record ?'));
                }

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
                            return callback();
                    }
                };

                log.info('OK')
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
};