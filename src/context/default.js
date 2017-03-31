/**
 * Created by igor on 27.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module);
const setupPickupParameters = require('./helper').setupPickupParameters;
const Call = require(__appRoot + '/router');
    
module.exports = (acr, conn, id, destinationNumber) => {
    const domainName = conn.channelData.getHeader('variable_domain_name');
    const callerIdNumber = conn.channelData.getHeader('Channel-Caller-ID-Number');

    conn.execute('unset', 'sip_h_call-info');

    if (callerIdNumber)
        conn.execute('hash', 'insert/spymap/${domain_name}-' + callerIdNumber + '/${uuid}');

    const findExtension = acr.db.getQuery('dialplan', 'findExtension');

    findExtension(destinationNumber, domainName, (err, extensionSchema) => {
        if (err) {
            return acr.closeConnection(conn, err);
        }

        if (extensionSchema)
            return internalCall(acr, conn, id, extensionSchema);

        const findDefault = acr.db.getQuery('dialplan', 'findDefault');
        findDefault(domainName, (err, defaultSchema) => {
            if (err) {
                return acr.closeConnection(conn, err);
            }

            return worldCall(acr, conn, id, defaultSchema);
        })
    });
};


function internalCall(acr, conn, id, shema) {
    if (!conn.channelData.getHeader('variable_webitel_direction')) {
        conn.execute('set', 'webitel_direction=internal');
        log.trace(`Set call ${id} direction internal`);
    }

    if (shema['fs_timezone']) {
        conn.execute('set', 'timezone=' + shema['fs_timezone']);
        log.trace(`Set call ${id} timezone ${shema['fs_timezone']}`);
    }

    new Call(conn, shema, acr);
    // TODO
}

function worldCall(acr, conn, id, shema) {

    // TODO
}