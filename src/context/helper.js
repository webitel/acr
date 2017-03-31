/**
 * Created by igor on 27.03.17.
 */

"use strict";
    
module.exports = {
    setupPickupParameters: (conn, userId, domain) => {
        if (!userId || !domain || !conn) {
            return log.error('Bad parameters setupPickupParameters');
        }

        conn.execute('export', 'dialed_extension=' + userId);
        conn.execute('hash', 'insert/' + domain + '-call_return/' + userId + '/${caller_id_number}');
        conn.execute('hash', 'insert/' + domain + '-last_dial_ext/' + userId + '/${uuid}');
        conn.execute('hash', 'insert/' + domain + '-last_dial_ext/global/${uuid}');
    }
};