/**
 * Created by i.navrotskyj on 10.02.2015.
 */
var log = require('../../lib/log')(module);

var execSyncApp = function (conn, app, data) {
    conn.setEventLock(true);
    log.trace('Execute app: %s, with data: %s', app, data || '');
};

module.exports = function (conn, userId, domainName) {
    userId = userId || '';
    domainName = domainName || '';

    conn.bgapi('user_exists id ' + userId + ' ' + domainName, function (res) {
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
            execSyncApp(conn, "lua", "RecordSession.lua");
            execSyncApp(conn, "bridge", "user/${destination_number}@${domain_name}");
            execSyncApp(conn, "answer");
            execSyncApp(conn, "sleep", "1500");
            execSyncApp(conn, "playback", "voicemail/vm-not_available_no_voicemail.wav");
            execSyncApp(conn, "hangup", "USER_NOT_REGISTERED");
        };
    });
};