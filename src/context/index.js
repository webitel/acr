/**
 * Created by igor on 27.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module);

const publicContext = require(__appRoot + '/context/public');
const defaultContext = require(__appRoot + '/context/default');
const dialerContext = require(__appRoot + '/context/dialer');

const PUBLIC_CONTEXT = 'public';

module.exports = (acr, conn, id) => {
    let lastExecuteDump;

    conn.on(`esl::event::CHANNEL_EXECUTE_COMPLETE::*`, e => lastExecuteDump = e);

    setSoundLang(conn);

    const context = conn.channelData.getHeader('Channel-Context');
    const dialerId = conn.channelData.getHeader('variable_dlr_queue');
    const destinationNumber = conn.channelData.getHeader('Channel-Destination-Number') ||
        conn.channelData.getHeader('Caller-Destination-Number') || conn.channelData.getHeader('variable_destination_number');

    acr.initGlobalVar(conn.channelData.getHeader('Core-UUID'), conn, e => {

        if (context === PUBLIC_CONTEXT) {
            log.debug(`Call ${id} from context public to: ${destinationNumber}`);
            publicContext(acr, conn, id, destinationNumber);
        } else if (dialerId) {
            log.debug(`Call ${id} from context dialer to: ${destinationNumber}`);
            dialerContext(acr, conn, id, destinationNumber);
        } else {
            log.debug(`Call ${id} from context default to: ${destinationNumber}`);
            defaultContext(acr, conn, id, destinationNumber)
        }

    });
};


function setSoundLang(conn) {
    if (conn.channelData.getHeader('variable_default_language') == 'ru') {
        conn.execute('set', 'sound_prefix=\/$${sounds_dir}\/ru\/RU\/elena');
    } else {
        conn.execute('set', 'sound_prefix=\/$${sounds_dir}\/en\/us\/callie');
    }
}