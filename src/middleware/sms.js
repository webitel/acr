/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

let Client = require('node-rest-client').Client,
    client = new Client(),
//EventEmitter2 = require('eventemitter2').EventEmitter2,
    log = require('./../lib/log')(module);

module.exports = function (parameters, rout, cb) {
    parameters = parameters || {};
    let login = parameters['login'] || '',
        pass = parameters['password'] || '',
        id = parameters['id'] || '1',
        sender = parameters['name'] || '',
        phone = parameters['phone'],
        sendTime = parameters['send_time'],
        encoding = parameters['encoding'],
        txt = parameters['message'] || '',
        _cb = false,
        isBulk = (phone instanceof Array && phone.length > 1)
        ;
    try {
        let xml = '<?xml version="1.0" encoding="UTF-8" ?>\n' +
                '<request method="send-sms" login="' + login + '" passw="' + pass + '">\n' +
                '\t<msg id="' + id + '"' +
                (!isBulk ? ' phone="' + phone.toString() + '"' : '') +
                ' sn="' + sender + '"' +
                (sendTime ? ' send_time="' + sendTime + '"' : '' ) +
                (encoding ? ' encoding="' + encoding + '"' : '') + '>' + txt + '</msg>\n'
            ;
        if (isBulk) {
            phone.forEach(function (number) {
                xml += '<phone number="' + number + '" />\n';
            });
        }

        xml += '</request>';
        log.trace('SMS XML: ' + xml);

        xml = rout._parseVariable(xml);

        let webArgs = {
            data: xml,
            headers: {
                'Content-Type': 'application/xml'
            },
            requestConfig: {
                timeout: 2000, //request timeout in milliseconds
                keepAlive: false //Enable/disable keep-alive functionalityidle socket.
            },
            responseConfig: {
                timeout: 2000 //response timeout
            }
        };

        let req = client.post('http://sms.barex.com.ua/websend/', webArgs, function (dataRequest) {
            if (rout.connection.socket) {
                rout.__setVar({
                    "setVar": "sendSmsResponse=" + dataRequest.toString()
                });
            }

            if (cb && !_cb) {
                // todo
                _cb = true;
                return cb();
            }

        });

        req.on('requestTimeout', function (req) {
            log.warn("request has expired");
            req.abort();
            if (cb && !_cb) {
                // todo
                _cb = true;
                return cb();
            }
        });

        req.on('error', function (err) {
            log.error(err.message);
            if (cb && !_cb) {
                // todo
                _cb = true;
                return cb();
            }
        });

        req.on('responseTimeout', function () {
            log.warn("response has expired");
        });
    } catch (e) {
        log.error(e);
    }
};