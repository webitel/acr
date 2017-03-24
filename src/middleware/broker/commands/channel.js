/**
 * Created by igor on 23.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module);

const Service = module.exports = {
    sttResponse: (msg = {}) => {
        const {callId, stt, setVar = "stt_response"} = msg;

        if (callId) {
            const conn = application.getConnection(callId);
            if (!conn)
                return log.debug(`Channel ${callId} is close`);

            if (conn && conn.__callRouter) {
                let transcript = '';
                if (stt && stt.result instanceof Array) {
                    const result = stt.result[0];
                    if (result && result.alternative instanceof Array && result.alternative.length > 0) {
                        transcript = result.alternative[0].transcript;
                    }
                }

                conn.__callRouter.__setVar({
                    "setVar": `${setVar}=${transcript}`
                }, () => {
                    if (conn.__callRouter._callbackStopApi instanceof Function) {
                        conn.__callRouter._callbackStopApi();
                    }
                });
                
            }
        } else {
            log.warn(`Bad stt message: `, msg);
        }
    }
};