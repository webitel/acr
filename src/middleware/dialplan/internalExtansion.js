/**
 * Created by i.navrotskyj on 10.02.2015.
 */
let log = require('../../lib/log')(module);

let execSyncApp = function (conn, app, data) {
    conn.setEventLock(true);
    conn.execute(app, data || '');
    log.trace('Execute app: %s, with data: %s', app, data || '');
};

module.exports = function (conn, userId, domainName) {
    userId = userId || '';
    domainName = domainName || '';

    conn.api('user_exists id ' + userId + ' ' + domainName, function (res) {
        try {
            if (res && res['body'] === "false") {
                execSyncApp(conn, "answer");
                execSyncApp(conn, "sleep", "1500");
                execSyncApp(conn, "playback", "ivr/ivr-you_have_dialed_an_invalid_extension.wav");
                execSyncApp(conn, "hangup", "UNALLOCATED_NUMBER");
            } else {
                execSyncApp(conn, "set", "continue_on_fail=true");
                execSyncApp(conn, "set", "hangup_after_bridge=true");
                execSyncApp(conn, "set", "effective_callee_id_number=${destination_number}");
                execSyncApp(conn, "set", "outbound_callee_id_number=${destination_number}");
                execSyncApp(conn, "set", "ringback=${ru-ring}");
                execSyncApp(conn, "set", "transfer_ringback=$${uk-ring}");
                execSyncApp(conn, "lua", "RecordSession.lua");
                conn.setEventLock(true);
                conn.execute("bridge", "user/${destination_number}@${domain_name}", function (res) {
                    try {
                        if (res && res.getHeader('variable_endpoint_disposition') !== 'ANSWER') {

                            execSyncApp(conn, "answer");
                            execSyncApp(conn, "sleep", "1500");
                            execSyncApp(conn, "playback", "voicemail/vm-not_available_no_voicemail.wav");
                            execSyncApp(conn, "hangup", "USER_NOT_REGISTERED");
                        }
                    } catch (e) {
                        log.warn(e.message);
                    } finally {
                        conn.disconnect();
                    }
                });
            }
        } catch (e) {
            log.error(e.message);
        }
    });
};