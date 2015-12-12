/**
 * Created by i.navrotskyj on 04.12.2015.
 */
'use strict';

var log = require('../../lib/log')(module),
    nodemailer = require('nodemailer'),
    smtpPool = require('nodemailer-smtp-pool'),
    helper = require('./helper')
;

const Provider = {
    "smtp": smtpPool
};

module.exports = function (CallRouter, APPLICATION_NAME) {
    // WTEL-270
    CallRouter.prototype.__sendEmail = function (app, cb) {
        let prop = app[APPLICATION_NAME];
        if (!prop) {
            log.error('Bad application parameters');
            return cb && cb();
        };

        let to = prop['to'],
            from = prop['from'],
            message = this._parseVariable(prop['message']),
            subject = prop['subject'] || '',
            domain = this.domain
        ;

        if (!to || !message) {
            log.error('Bad email parameters');
            return cb && cb();
        };

        let mailOption = {
            to: to,
            subject: subject,
            html: message
        };

        helper
            .getSettings(
                domain,
                (err, res) => {
                    if (err)
                        return log.error(err);

                    if (!res) {
                        return log.error('Not found parameters in %s.', domain);
                    };

                    if (typeof Provider[res.provider] != 'function') {
                        return log.error('Bad provider name in %s.', domain);
                    };

                    if (!res || !res['options']) {
                        return log.error("Not settings EMail provider from domain " + domain);
                    };

                    mailOption.from = from || res.from || '';

                    try {
                        let transport = nodemailer.createTransport(Provider[res.provider](res['options']));
                        transport.sendMail(
                            mailOption,
                            (err, res) => {
                                if (err)
                                    return log.error(err.message);

                                return log.debug(res);
                            }
                        );
                    } catch (e) {
                        log.error(e);
                    };
                }
            );

        return cb && cb();

    };
};