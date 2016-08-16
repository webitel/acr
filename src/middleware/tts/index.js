/**
 * Created by igor on 11.08.16.
 */

"use strict";
const aws = require('./aws4'),
    crypto = require('crypto'),
    log = require('../../lib/log')(module)
    ;

const PROVIDER = {
    "ivona": (router, config, cb) => {
        if (!config.ivonaAccessKey || !config.ivonaSecretKey || !config.text) {
            log.error(`Bad config parameters Ivona: ivonaAccessKey, ivonaSecretKey, text is required`);
            return cb && cb();
        }

        let keys = {
            accessKeyId : config.ivonaAccessKey,
            secretAccessKey : config.ivonaSecretKey
        };

        let text = router._parseVariable(config.text),
            voice = config.voice || {};


        let query = `Input.Data=${encodeURIComponent(text)}&Input.Type=${encodeURIComponent('text/plain')}` +
            `&OutputFormat.Codec=MP3&OutputFormat.SampleRate=22050&Parameters.Rate=slow`
            ;


        if (voice.gender) {
            query += `&Voice.Gender=${voice.gender}`
        }

        if (voice.language) {
            query += `&Voice.Language=${voice.language}`
        }

        if (voice.name) {
            query += `&Voice.Name=${voice.name}`
        }
        
        let request = {
            path: `/CreateSpeech?${query}`,
            host: 'tts.eu-west-1.ivonacloud.com',
            service: 'tts',
            method: 'GET',
            region: config.region || 'eu-west-1',
            accessKeyId : config.ivonaAccessKey,
            secretAccessKey : config.ivonaSecretKey,
            body: ''
        };
        let signedCredentials = aws.signUrl(request, keys);

        router.execApp({
            "app": 'playback',
            "data": 'shout://tts.eu-west-1.ivonacloud.com:443' + signedCredentials
        }, cb);

    }
};

module.exports = (CallRouter, appName) => {
    CallRouter.prototype['__' + appName] = function (app, cb) {
        let prop = app[appName];
        if (PROVIDER.hasOwnProperty(prop.provider)) {
            return PROVIDER[prop.provider](this, prop, cb);
        }

        log.error(`Provider not found.`);
        return cb && cb();
    }
};