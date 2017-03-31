/**
 * Created by igor on 27.03.17.
 */

"use strict";
    
module.exports = {
    api: (str = "") => {
        const _json = {};
        let _param;

        if (str) {
            str.split('\n').forEach(function (str) {
                _param = str.split('=');
                if (_param[0] == '') return;
                _json[_param[0]] = _param[1];
            });
        }

        return _json;
    }
};