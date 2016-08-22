/**
 * Created by igor on 11.08.16.
 */

"use strict";
const log = require('../../lib/log')(module)
    ;

const PROVIDER = {
    "microsoft": (router, config, cb) => {

        if (!config.accessKey1 || !config.accessKey2 || !config.appId || !config.text) {
            log.error(`Bad config parameters microsoft: accessKey1, accessKey2, appId, text is required`);
            return cb && cb();
        }

        let text = router._parseVariable(config.text),
            voice = config.voice || {},
            playbackProp = copyTTSOptionToPlayback(config);

        let query = `text=${encodeURIComponent(text)}`;

        if (voice.gender) {
            query += `&gender=${voice.gender}`
        }

        if (voice.language) {
            query += `&language=${voice.language}`
        }

        if (voice.name) {
            query += `&name=${voice.name}`
        }

        query += `&key1=${config.accessKey1}&key2=${config.accessKey2}&appId=${config.appId}&.wav`;

        playbackProp.name = 'http_cache://$${cdr_url}/sys/tts/microsoft?' + query;
        playbackProp.type = 'local';

        router.__playback({
            playback: playbackProp
        }, cb)

    },
    "ivona": (router, config, cb) => {
        if (!config.accessKey || !config.accessToken || !config.text) {
            log.error(`Bad config parameters Ivona: ivonaAccessKey, ivonaSecretKey, text is required`);
            return cb && cb();
        }

        let text = router._parseVariable(config.text),
            voice = config.voice || {},
            playbackProp = copyTTSOptionToPlayback(config);


        let query = `text=${encodeURIComponent(text)}`;


        if (voice.gender) {
            query += `&gender=${voice.gender}`
        }

        if (voice.language) {
            query += `&language=${voice.language}`
        }

        if (voice.name) {
            query += `&name=${voice.name}`
        }

        query += `&key=${encodeURIComponent(config.accessKey)}&token=${encodeURIComponent(config.accessToken)}`;

        playbackProp.name = `${router.getGlbVar('cdr_url').replace(/https?/, 'shout')}/sys/tts/ivona?${query}`;
        playbackProp.type = 'local';

        router.__playback({
            playback: playbackProp
        }, cb)

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

const _PLAYBACK_PROPS = ['getDigits', 'broadcast', 'terminator'];

function copyTTSOptionToPlayback(ttsProp) {
    let playbackProp = {};

    for (let key in ttsProp) {
        if (ttsProp.hasOwnProperty(key) && ~_PLAYBACK_PROPS.indexOf(key)) {
            playbackProp[key] = ttsProp[key]
        }
    }

    return playbackProp;
}