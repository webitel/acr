/**
 * Created by igor on 11.08.16.
 */

"use strict";
const log = require('../../lib/log')(module)
    ;

const PROVIDER = {
    "microsoft": (router, config, cb) => {
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

        const {rate, format} = getCodecSettings(router.getChnVar('write_rate'));
        query += `&rate=${rate}&format=${format}`;

        if (config.accessKey && config.accessToken)
            query += `&accessKey=${config.accessKey}&accessToken=${config.accessToken}`;

        playbackProp.name = `${router.getGlbVar('cdr_url').replace(/https?/, format === 'mp3' ? 'shout': 'http_cache')}/sys/tts/microsoft?${query}`;
        playbackProp.type = 'local';

        router.__playback({
            playback: playbackProp
        }, cb)

    },
    "ivona": (router, config, cb) => {

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

        if (config.accessKey && config.accessToken)
            query += `&accessKey=${encodeURIComponent(config.accessKey)}&accessToken=${encodeURIComponent(config.accessToken)}`;

        const {rate, format} = getCodecSettings(router.getChnVar('write_rate'));
        query += `&rate=${rate}&format=${format}`;

        playbackProp.name = `${router.getGlbVar('cdr_url').replace(/https?/, format === 'mp3' ? 'shout': 'http_cache')}/sys/tts/ivona?${query}`;
        playbackProp.type = 'local';

        router.__playback({
            playback: playbackProp
        }, cb)

    },
    "polly": (router, config, cb) => {

        let text = router._parseVariable(config.text),
            voice = config.voice,
            playbackProp = copyTTSOptionToPlayback(config);


        let query = `text=${encodeURIComponent(text)}`;

        if (voice) {
            query += `&voice=${voice}`
        }

        if (config.accessKey && config.accessToken)
            query += `&accessKey=${encodeURIComponent(config.accessKey)}&accessToken=${encodeURIComponent(config.accessToken)}`;

        const {rate, format} = getCodecSettings(router.getChnVar('write_rate'));
        query += `&rate=${rate}&format=${format}`;

        playbackProp.name = `${router.getGlbVar('cdr_url').replace(/https?/, format === 'mp3' ? 'shout': 'http_cache')}/sys/tts/polly?${query}`;
        playbackProp.type = 'local';

        router.__playback({
            playback: playbackProp
        }, cb)

    },
    "default": (router, config, cb) => {
        let text = router._parseVariable(config.text),
            playbackProp = copyTTSOptionToPlayback(config);

        let query = `text=${encodeURIComponent(text)}`;

        for (let key in config) {
            if (config.hasOwnProperty(key) && !~_PLAYBACK_PROPS.indexOf(key) && key !== 'text') {
                if (config[key] instanceof Object) {
                    for (let keyObject in config[key]) {
                        query += `&${keyObject}=${config[key][keyObject]}`
                    }
                } else {
                    query += `&${key}=${config[key]}`
                }
            }
        }

        const {rate, format} = getCodecSettings(router.getChnVar('write_rate'));
        query += `&rate=${rate}&format=${format}`;

        playbackProp.name = `${router.getGlbVar('cdr_url').replace(/https?/, format === 'mp3' ? 'shout': 'http_cache')}/sys/tts/default?${query}`;
        playbackProp.type = 'local';
        
        router.__playback({
            playback: playbackProp
        }, cb)
    }
};

module.exports = (CallRouter, appName) => {
    CallRouter.prototype['__' + appName] = function (app, cb) {
        let prop = app[appName];

        if (!prop.text) {
            log.error(`Bad config parameters tts: text is required`);
            return cb && cb();
        }

        if (PROVIDER.hasOwnProperty(prop.provider)) {
            return PROVIDER[prop.provider](this, prop, cb);
        } else {
            return PROVIDER.default(this, prop, cb);
        }
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

function getCodecSettings(writeRate) {
    let rate = +writeRate || 8000;
    let format = 'mp3';

    if (rate === 8000 || rate === 16000) {
        format = '.wav'
    } else if (rate > 22050) {
        rate = 22050;
    }
    return {rate, format};
}