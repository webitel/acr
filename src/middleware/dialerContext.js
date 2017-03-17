/**
 * Created by igor on 14.05.16.
 */

'use strict';


let log = require('./../lib/log')(module),
    dialplan = require('./dialplan'),
    DEFAULT_HANGUP_CAUSE = require('../const').DEFAULT_HANGUP_CAUSE,
    CallRouter = require('./callRouter');

module.exports = function (conn, destinationNumber, globalVariable) {
    let domainName = conn.channelData.getHeader('variable_domain_name');

    dialplan.findDialerDialplan(destinationNumber, domainName, (err, res) => {
        if (err) {
            log.error(err.message);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        }

        if (!res || !(res._cf instanceof Array)) {
            log.error(`Not found dialer ${destinationNumber} context`);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        }

        if (!domainName) {
            log.error(`Not found domain ${domainName} -> ${destinationNumber} context`);
            conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            return
        }

        let callflow;
        if (res.amd && res.amd.enabled) {
            callflow = [].concat(getAmdSection(null, res.amd), res._cf, getFooter())
        } else {
            callflow = [].concat(res._cf, getFooter())
        }

        // TODO caller ?
        let dn = conn.channelData.getHeader('Caller-Caller-ID-Number') || destinationNumber,
            uuid = conn.channelData.getHeader('variable_uuid'),
            answeredTime = conn.channelData.getHeader('Caller-Channel-Answered-Time')
            ;
        conn.execute('set', 'webitel_direction=dialer');


        let _router = new CallRouter(conn, {
            "globalVar": globalVariable,
            "desNumber": dn,
            "chnNumber": dn,
            "timeOffset": null,
            "versionSchema": 2,
            "domain": domainName
        });

        let exec = function () {
            try {
                log.trace('Exec: %s', dn);
                _router.run(callflow);
            } catch (e) {
                log.error(e.message);
                //TODO узнать что ответить на ошибку
                conn.execute('hangup', DEFAULT_HANGUP_CAUSE);
            }
        };

        // conn.subscribe('CUSTOM avmd::beep');
        //
        // conn.once(`esl::event::CUSTOM::${uuid}`, (e) => {
        //     console.log(e.serialize());
        // });
        //
        // conn.execute(`avmd_start`, "simplified_estimation=0,inbound_channel=1,outbound_channel=1,sample_n_continuous_streak=13,sample_n_to_skip=18,debug=1,report_status=0,fast_math=1", (res) => {
        //     console.log(res.serialize())
        // });

        if (+answeredTime > 0) {
            log.trace(`Channel ${uuid}  answered ${answeredTime}`);
            exec();
        } else {
            log.trace(`Channel not answered, subscribe CHANNEL_ANSWER`);
            conn.subscribe('CHANNEL_ANSWER');
            conn.once('esl::event::CHANNEL_ANSWER::*', () => {
                log.trace(`On CHANNEL_ANSWER ${uuid}`);
                exec();
            });
        }
    });
};

const AMD_PARAMS = ["maximumWordLength", "maximumNumberOfWords", "betweenWordsSilence", "minWordLength",
    "totalAnalysisTime", "silenceThreshold", "afterGreetingSilence", "greeting", "initialSilence"];
function getAmdSection(channel, amdConfig = {}) {
    const amdParams = {};
    for (let param of AMD_PARAMS) {
        if (amdConfig.hasOwnProperty(param))
            amdParams[param] = amdConfig[param]
    }

    return [
        {
            "setVar": "ignore_early_media=true"
        },
        {
            "answer": ""
        },
        {
            "amd": amdParams
        },
        {
            "if": {
                "expression": "${amd_result} !== 'HUMAN'",
                "sysExpression" : "sys.getChnVar(\"amd_result\") !== 'HUMAN'",
                "then": [
                    {
                        "hangup": "USER_BUSY"
                    },
                    {
                        "break": true
                    }
                ]
            }
        }
    ]
}


function getFooter() {
    return [
        {
            "hangup": ""
        }
    ]
}