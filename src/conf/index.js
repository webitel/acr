/**
 * Created by igor on 27.03.17.
 */

"use strict";

const nconf = require('nconf');
const path = require('path');

nconf.argv()
    .env()
    .file({
        file: path.join(__dirname, 'config.json')
    });

module.exports = nconf;